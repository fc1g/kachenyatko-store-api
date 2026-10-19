package telegram

import (
	"context"
	"fmt"

	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/go-telegram/bot"
)

type Bot struct {
	bot    *bot.Bot
	chatID int64
}

func NewBot(cfg config.Telegram) (*Bot, error) {
	b, err := bot.New(cfg.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("initialize telegram bot: %w", err)
	}

	return &Bot{
		bot:    b,
		chatID: cfg.ChatID,
	}, nil
}

func (b *Bot) SendMessage(ctx context.Context, text string) error {
	_, err := b.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: b.chatID,
		Text:   text,
	})
	if err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}

	return nil
}
