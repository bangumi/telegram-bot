import asyncio
import http
import logging
import secrets
import sys
import traceback

import httpx
import msgspec.json
import redis.asyncio as redis
import starlette
import starlette.applications
import telegram as tg
import telegram.ext
import uvicorn
import yarl
from loguru import logger
from starlette.exceptions import HTTPException
from starlette.requests import Request
from starlette.responses import PlainTextResponse, RedirectResponse, Response

import db
from db import DB, create_db
from lib import config

logging.basicConfig(
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO
)
logging.getLogger("httpx").setLevel(logging.WARN)


class RedisOAuthState(msgspec.Struct):
    chat_id: int


class BangumiOAuthResponse(msgspec.Struct):
    user_id: int


def state_to_redis_key(state: str):
    return f"tg-bot-oauth:{state}"


class TelegramApplication:
    def __init__(self, redis: redis.Redis, db_client: db.DB):
        application = tg.ext.Application.builder()

        if sys.platform == "win32":
            proxy_url = "http://127.0.0.1:7890"
            application = application.proxy_url(proxy_url).get_updates_proxy_url(
                proxy_url
            )

        application = application.token(config.TELEGRAM_BOT_TOKEN).build()

        # on different commands - answer in Telegram
        application.add_handler(tg.ext.CommandHandler("start", self.start_command))

        application.add_error_handler(self.error_handler)
        self.app = application
        self.redis = redis
        self.db = db_client

    @logger.catch
    async def start_command(
            self, update: tg.Update, context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        """Send a message when the command /help is issued."""
        logger.info("help command")
        if user := await self.db.is_authorized_user(chat_id=update.effective_chat.id):
            await update.message.reply_text(f"你已经作为用户 {user.user_id} 成功进行认证")
            return

        token = secrets.token_urlsafe(32)
        await self.redis.set(
            state_to_redis_key(token),
            msgspec.json.encode(RedisOAuthState(chat_id=update.effective_chat.id)),
            ex=60,
        )
        reply_markup = tg.InlineKeyboardMarkup(
            [
                [
                    tg.InlineKeyboardButton(
                        "认证 bangumi 账号",
                        url=f"{config.EXTERNAL_HTTP_ADDRESS}/redirect?state={token}",
                    )
                ]
            ]
        )
        await update.message.reply_text("请在60s内进行认证", reply_markup=reply_markup)

    @logger.catch
    async def error_handler(
            self, update: object, context: tg.ext.ContextTypes.DEFAULT_TYPE
    ) -> None:
        """Log the error and send a telegram message to notify the developer."""
        # Log the error before we do anything else, so we can see it even if something breaks.
        logger.error("Exception while handling an update:", exc_info=context.error)

        # traceback.format_exception returns the usual python message about an exception, but as a
        # list of strings rather than a single string, so we have to join them together.
        tb_list = traceback.format_exception(
            None, context.error, context.error.__traceback__
        )
        tb_string = "".join(tb_list)

        # Build the message with some markup and additional information about what happened.
        # You might need to add some logic to deal with messages longer than the 4096 character limit.
        update_str = update.to_dict() if isinstance(update, tg.Update) else str(update)
        print(tb_string)
        print(context.user_data)
        print(context.chat_data)
        print(update_str)

    async def _save_user_chat_map(self, user_id: int, chat: int):
        pass


class OAuthHTTPServer:
    def __init__(self, redis: redis.Redis, db: DB, bot: tg.ext.Application):
        self.app = starlette.applications.Starlette()
        self.app.add_route("/", self.index_path, ["GET"])
        self.app.add_route("/redirect", self.oauth_redirect, ["GET"])
        self.app.add_route("/callback", self.oauth_callback, ["GET"])
        self.redirect_url = str(config.EXTERNAL_HTTP_ADDRESS.with_path("/callback"))

        self.redis = redis

        self.http_client = httpx.AsyncClient()
        self.db = db
        self.tg = bot

    async def index_path(self, request: Request) -> Response:
        return Response("index page")

    async def oauth_redirect(self, request: Request) -> Response:
        return RedirectResponse(
            str(
                yarl.URL.build(
                    scheme="https",
                    host="bgm.tv",
                    path="/oauth/authorize",
                    query={
                        "client_id": config.BANGUMI_APP_ID,
                        "response_type": "code",
                        "redirect_uri": self.redirect_url,
                        "state": request.query_params.get("state"),
                    },
                )
            )
        )

    async def oauth_callback(self, request: Request) -> Response:
        code = request.query_params.get("code")
        state = request.query_params.get("state")
        if not code or not state:
            raise HTTPException(http.HTTPStatus.BAD_REQUEST, "非法请求，请使用telegram重新获取认证链接")
        redis_state_value_raw = await self.redis.get(state_to_redis_key(state))
        if redis_state_value_raw is None:
            raise HTTPException(http.HTTPStatus.BAD_REQUEST, "非法请求，请使用telegram重新获取认证链接")

        redis_state = msgspec.json.decode(redis_state_value_raw, type=RedisOAuthState)

        resp = await self.http_client.post(
            "https://bgm.tv/oauth/access_token",
            data={
                "client_id": config.BANGUMI_APP_ID,
                "client_secret": config.BANGUMI_APP_SECRET,
                "grant_type": "authorization_code",
                "code": code,
                "redirect_uri": self.redirect_url,
            },
        )
        if resp.status_code >= 300:
            logger.error("bad oauth response", data=resp.json())
            raise HTTPException(http.HTTPStatus.BAD_GATEWAY, "请尝试重新认证")
        data = msgspec.json.decode(resp.text, type=BangumiOAuthResponse)

        await self.db.insert_chat_bangumi_map(
            user_id=data.user_id, chat_id=redis_state.chat_id
        )
        await self.tg.bot.send_message(
            chat_id=redis_state.chat_id, text=f"已经成功关联用户 {data.user_id}"
        )

        return PlainTextResponse("你已经成功认证，请关闭页面返回 telegram")

    async def start(self):
        logger.info("start http server")
        port = config.HTTP_PORT or config.EXTERNAL_HTTP_ADDRESS.port or 4098
        server = uvicorn.Server(uvicorn.Config(app=self.app, port=port))
        logger.info("http server listen on port={}", port)
        await server.serve()


async def start() -> None:
    redis_client = redis.from_url(str(config.REDIS_DSN))

    db = await create_db()

    app = TelegramApplication(redis=redis_client, db_client=db)
    async with app.app:
        await app.app.initialize()
        await app.app.start()
        await app.app.updater.start_polling(allowed_updates=tg.Update.ALL_TYPES)

        logger.info("telegram bot start")

        await OAuthHTTPServer(redis=redis_client, db=db, bot=app.app).start()


def main() -> None:
    asyncio.run(start())


if __name__ == "__main__":
    main()
