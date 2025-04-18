package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
)

const oauthURL = "https://next.bgm.tv/oauth/authorize"

func (h *handler) ListenAndServe() error {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Get("/", http.RedirectHandler(fmt.Sprintf("https://t.me/%s", h.botUser.Username), http.StatusFound).ServeHTTP)
	mux.Get("/callback", h.handleOAuthCallback)
	mux.Get("/redirect", h.oauthRedirect)
	return http.ListenAndServe(fmt.Sprintf(":%d", h.config.Port), mux)
}

func (h *handler) oauthRedirect(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "请求无效，请在 telegram 中重新认证", http.StatusBadRequest)
		return
	}

	query := url.Values{}
	query.Add("client_id", h.config.BangumiAppId)
	query.Add("response_type", "code")
	query.Add("redirect_uri", fmt.Sprintf("%s/callback", h.config.ExternalHttpAddress))
	query.Add("state", state)

	http.Redirect(w, r, oauthURL+"?"+query.Encode(), http.StatusFound)
}

// handleOAuthCallback processes the OAuth callback request
func (h *handler) handleOAuthCallback(w http.ResponseWriter, req *http.Request) {
	// Get query parameters
	q := req.URL.Query()
	code := q.Get("code")
	state := q.Get("state")

	if code == "" || state == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	var r OAuthAccessTokenResponse

	resp, err := h.client.R().
		SetFormData(map[string]string{
			"client_id":     h.config.BangumiAppId,
			"client_secret": h.config.BangumiAppSecret,
			"grant_type":    "authorization_code",
			"code":          code,
			"redirect_uri":  h.config.ExternalHttpAddress + "/callback",
		}).
		SetResult(&r).
		Post("https://next.bgm.tv/oauth/access_token")

	if err != nil {
		panic(err)
	}

	if resp.StatusCode() >= 300 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, "请求错误", http.StatusBadRequest)
		return
	}

	v, err := h.redis.Do(req.Context(), h.redis.B().Get().Key(redisStateKey(state)).Build()).AsBytes()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("请重新认证"))
		}
		return
	}

	var redisState RedisOAuthState
	_ = json.Unmarshal(v, &redisState)

	_, err = h.pg.ExecContext(req.Context(), `
	INSERT INTO telegram_notify_chat(chat_id, user_id, disabled) VALUES ($1, $2, 0)
	on conflict (chat_id, user_id) do update
	set disabled = 0
	`,
		redisState.ChatID,
		r.UserID,
	)

	if err != nil {
		log.Err(err).Msg("failed to save data to pg")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("发生未知错误"))
		return
	}

	_, err = h.bot.SendMessage(req.Context(), tu.Message(
		tu.ID(redisState.ChatID),
		fmt.Sprintf("已经成功关联用户 %s", r.UserID),
	))

	if err != nil {
		log.Err(err).Msg("failed to send message to user")
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("你已经成功认证，请关闭页面返回 telegram"))

}

type OAuthAccessTokenResponse struct {
	UserID string `json:"user_id"`
}

// RedisOAuthState represents the OAuth state stored in Redis
type RedisOAuthState struct {
	ChatID int64 `json:"chat_id"`
}

func redisStateKey(state string) string {
	return "tg-bot-oauth:" + state
}
