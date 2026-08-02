FROM gcr.io/distroless/static@sha256:9197324ba51d9cd071af8505989365c006adf9d6d2067eada25aef00abbb5278

ENTRYPOINT ["/app/telegram-bot"]

COPY /dist/telegram-bot /app/telegram-bot
