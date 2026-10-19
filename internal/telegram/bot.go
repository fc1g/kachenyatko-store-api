package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/fc1g/kachenyatko-store-api/internal/config"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Bot struct {
	bot         *bot.Bot
	chatID      int64
	sendTimeout time.Duration
}

func NewBot(cfg config.Telegram) (*Bot, error) {
	b, err := bot.New(cfg.AccessToken, bot.WithSkipGetMe())
	if err != nil {
		return nil, fmt.Errorf("initialize telegram bot: %w", err)
	}

	return &Bot{
		bot:         b,
		chatID:      cfg.ChatID,
		sendTimeout: cfg.SendTimeout,
	}, nil
}

func (b *Bot) SendMessage(ctx context.Context, text string) error {
	ctx, cancel := context.WithTimeout(ctx, b.sendTimeout)
	defer cancel()

	_, err := b.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    b.chatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}

	return nil
}
