package app

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"github.com/nktauserum/md-2320/internal/config"
	"github.com/nktauserum/md-2320/internal/workers"
)

const (
	updateInterval = 5 * time.Second
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
		ParseMode: telego.ModeMarkdown,
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
	var lastPercentage string = "0%"
	var lastUpdateTime time.Time

	go h.w["youtube"](message.Text, messages)

	for msg := range messages {
		switch msg.Type {
		case workers.MessageTypeInfo:
			logBuilder.WriteString("\n" + msg.Content)
		case workers.MessageTypeError:
			logBuilder.WriteString("\n**" + msg.Content + "**")
		case workers.MessageTypeProgress:
			jsonStr := strings.ReplaceAll(msg.Content, "'", "\"")

			var progress map[string]string
			err := json.Unmarshal([]byte(jsonStr), &progress)
			if err != nil {
				if logBuilder.Len() > 0 {
					logBuilder.WriteString("\n")
				}
				logBuilder.WriteString("**error parsing json string: " + err.Error() + "**")
				continue
			}

			lastPercentage = strings.TrimSpace(progress["progress_percentage"])
		}

		now := time.Now()
		if now.Sub(lastUpdateTime) >= updateInterval || msg.Type == workers.MessageTypeError {
			text := fmt.Sprintf("%s\n\n**progress: %s**", logBuilder.String(), lastPercentage)
			update_msg(ctx, tg_msg, text)
			lastUpdateTime = now
		}
	}

	text := fmt.Sprintf("%s\n\nprogress: %s (completed)", logBuilder.String(), lastPercentage)
	update_msg(ctx, tg_msg, text)

	return nil
}
