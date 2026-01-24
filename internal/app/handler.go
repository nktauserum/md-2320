package app

import (
	"slices"

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

func send_text(ctx *th.Context, chat_id telego.ChatID, text string) {
	ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    chat_id,
		Text:      text,
		ParseMode: telego.ModeMarkdown,
	})
}

func (h *Handler) HandleRequest(ctx *th.Context, message telego.Message) error {
	if !h.is_authorized(message.From.ID) {
		send_text(ctx, message.Chat.ChatID(),
			"You're not authorized to perform this operation. Get your permissions from the main technopriest.")
		return nil
	}

	messages := make(chan string)

	go h.w["youtube"](message.Text, messages)
	for msg := range messages {
		send_text(ctx, message.Chat.ChatID(), msg)
	}

	return nil
}
