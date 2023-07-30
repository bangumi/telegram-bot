# bgm.tv Telegram bot

监听 debezium + kafka 的数据库变更事件，在 telegram 上提醒用户新通知。

完全基于 asyncio 。 使用 PostgresSQL 作为数据库，[asyncpg](https://magicstack.github.io/asyncpg/current/) 作为 driver ，未使用 ORM。

欢迎 PR 添加新功能。
