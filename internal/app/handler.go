package app

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"github.com/nktauserum/md-2320/internal/config"
	"github.com/nktauserum/md-2320/internal/workers"
)

type Handler struct {
	c *config.Config
	w map[string]workers.Worker
}

func NewHandler(c *config.Config, workers map[string]workers.Worker) *Handler {
	return &Handler{c, workers}
}

func (h *Handler) is_authorized(user_id int64) bool {
	return slices.Contains(h.c.AUTHORIZED_USERS, user_id)
}

func send_text(ctx *th.Context, chat_id telego.ChatID, text string) (*telego.Message, error) {
	return ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    chat_id,
		Text:      text,
		ParseMode: telego.ModeMarkdown,
	})
}

func update_msg(ctx *th.Context, msg *telego.Message, text string) {
	ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    msg.Chat.ChatID(),
		MessageID: msg.MessageID,
		Text:      text,
	})
}

func (h *Handler) HandleRequest(ctx *th.Context, message telego.Message) error {
	if !h.is_authorized(message.From.ID) {
		send_text(ctx, message.Chat.ChatID(),
			"You're not authorized to perform this operation. Get your permissions from the main technopriest.")
		return nil
	}

	tg_msg, _ := send_text(ctx, message.Chat.ChatID(), "processing your request...")

	messages := make(chan workers.Message)

	var logBuilder strings.Builder
	go h.w["youtube"](message.Text, messages)

	for msg := range messages {
		var percentage string = "0%"

		switch msg.Type {
		case workers.MessageTypeInfo:
			logBuilder.WriteString(msg.Content)
		case workers.MessageTypeError:
			logBuilder.WriteString("**" + msg.Content + "**")
		case workers.MessageTypeProgress:
			jsonStr := strings.ReplaceAll(msg.Content, "'", "\"")

			var progress map[string]string
			err := json.Unmarshal([]byte(jsonStr), &progress)
			if err != nil {
				messages <- workers.Message{Type: workers.MessageTypeError, Content: ("error parsing json string: " + err.Error())}
				continue
			}

			percentage = strings.TrimSpace(progress["progress_percentage"])
		}

		update_msg(ctx, tg_msg, fmt.Sprintf("%s\n\nprogress: %s", logBuilder.String(), percentage))
	}

	return nil
}
