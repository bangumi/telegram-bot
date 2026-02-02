FROM gcr.io/distroless/static@sha256:972618ca78034aaddc55864342014a96b85108c607372f7cbd0dbd1361f1d841

ENTRYPOINT ["/app/telegram-bot"]

COPY /dist/telegram-bot /app/telegram-bot
