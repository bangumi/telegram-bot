FROM gcr.io/distroless/static@sha256:d9f9472a8f4541368192d714a995eb1a99bab1f7071fc8bde261d7eda3b667d8

ENTRYPOINT ["/app/telegram-bot"]

COPY /dist/telegram-bot /app/telegram-bot
