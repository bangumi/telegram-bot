FROM gcr.io/distroless/static@sha256:b7b9a6953e7bed6baaf37329331051d7bdc1b99c885f6dbeb72d75b1baad54f9

ENTRYPOINT ["/app/telegram-bot"]

COPY /dist/telegram-bot /app/telegram-bot
