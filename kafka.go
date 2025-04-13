package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func (h *handler) processKafkaMessage() error {
	k := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{h.config.KafkaBroker},
		GroupID: "tg-notify-bot",
		GroupTopics: []string{
			"debezium.chii.bangumi.chii_pms",
			"debezium.chii.bangumi.chii_notify",
		},
	})

	for {
		msg, err := k.ReadMessage(context.Background())
		if err != nil {
			log.Err(err).Msg("failed to read kafka message")
			continue
		}

		// fake event generated by debezium, ignore it
		if len(msg.Value) == 0 {
			continue
		}

		switch msg.Topic {
		case "debezium.chii.bangumi.chii_pms":
			err = h.handlePM(msg)
		case "debezium.chii.bangumi.chii_notify":
			err = h.handleNotify(msg)
		}

		if err != nil {
			log.Err(err).Msg("failed to parse kafka message")
		}
	}
}
func (h *handler) handlePM(msg kafka.Message) error {
	if len(msg.Value) == 0 {
		return nil
	}

	var dv DebeziumValue
	if err := json.Unmarshal(msg.Value, &dv); err != nil {
		log.Err(err).Bytes("value", msg.Value).Msg("failed to unmarshal debezium value for PM")
		return err // Return error if unmarshalling fails
	}

	// Ignore delete or update operations, only handle create
	if dv.Op != OpCreate {
		return nil
	}

	// Ignore events without payload
	if len(dv.After) == 0 {
		return nil
	}

	// Ignore old messages (older than 2 minutes)
	// Use UnixMilli() for millisecond timestamp comparison
	if time.Now().UnixMilli()-dv.Source.TsMs > 120*1000 {
		log.Debug().Int64("ts_ms", dv.Source.TsMs).Msg("Skipping old PM message")
		return nil
	}

	var pm ChiiPm
	if err := json.Unmarshal(dv.After, &pm); err != nil {
		log.Err(err).RawJSON("after", dv.After).Msg("failed to unmarshal chii_pms after value")
		return err // Return error if unmarshalling fails
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Get the chats for the recipient user
	chats, err := h.getChats(ctx, pm.MsgRid)
	if err != nil {
		// Log error but don't stop processing other messages
		log.Err(err).Int64("user_id", pm.MsgRid).Msg("failed to get chats for PM recipient")
		return err // Return error if getting chats fails
	}
	if len(chats) == 0 {
		return nil // No chats to send to
	}

	// Get sender user info
	fromUser, err := h.getUserInfo(ctx, pm.MsgSid)
	if err != nil {
		return err
	}

	pmURL := fmt.Sprintf("https://bgm.tv/pm/view/%d.chii", pm.MsgId)

	text := "收到来自 <b>" + html.EscapeString(fromUser.Nickname) + "</b> 的新私信"
	text = text + "\n\n" + pmURL

	for _, chatID := range chats {
		message := tu.Message(tu.ID(chatID), text).WithParseMode(telego.ModeHTML)
		_, _ = h.bot.SendMessage(ctx, message)
	}
	return nil
}

func (h *handler) handleNotify(msg kafka.Message) error {
	var dv DebeziumValue
	_ = json.Unmarshal(msg.Value, &dv)
	if len(dv.After) == 0 {
		return nil
	}

	if dv.Source.TsMs-time.Now().UnixMicro() < 60*2*1000 {
		// skip notification older than 2 min
		return nil
	}

	if dv.Op != OpCreate {
		return nil
	}

	var notify ChiiNotify
	_ = json.Unmarshal(dv.After, &notify)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Get the chats for this user
	chats, err := h.getChats(ctx, notify.Uid)
	if err != nil {
		return err
	}
	if len(chats) == 0 {
		return nil
	}

	cfg, hasValue := notifyConfigs[notify.Type]
	if !hasValue {
		log.Warn().Msgf("missing config for type %d", notify.Type)
		return nil
	}

	var field ChiiNotifyField
	err = h.mysql.QueryRowxContext(ctx, "SELECT ntf_id,ntf_hash,ntf_rid,ntf_title from chii_notify_field where ntf_id = %s limit 1",
		notify.Mid).StructScan(&field)
	if err != nil {
		return err
	}

	fromUser, err := h.getUserInfo(ctx, notify.FromUid)
	if err != nil {
		return err
	}

	// Construct URL
	url := strings.TrimRight(cfg.URL, "/") + "/" + strconv.FormatInt(notify.Mid, 10)
	if notify.Mid > 0 {
		url += cfg.Anchor + strconv.FormatInt(notify.Mid, 10)
	}

	var buf = bytes.NewBuffer(nil)
	err = cfg.Temp.Execute(buf, TmplData{
		Title:        field.NtfTitle,
		FromNickname: fromUser.Nickname,
	})
	if err != nil {
		return err
	}

	var text = buf.String() + "\n\n" + url

	log.Info().Int64("user_id", notify.Uid).Msg("should send message for notification")

	// Send message to all chats
	for _, chatID := range chats {
		if _, err := h.bot.SendMessage(ctx, tu.Message(tu.ID(chatID), text).WithParseMode(telego.ModeHTML)); err != nil {
			log.Err(err).Int64("chat_id", chatID).Msg("failed to send notification")
		}
	}

	return nil
}
