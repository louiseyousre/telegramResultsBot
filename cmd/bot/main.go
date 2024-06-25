package main

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, _ := bot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)

	// call methods.SetWebhook if needed

	go b.StartWebhook(ctx)

	log.Print("Starting the bot server...")

	err := http.ListenAndServe(":2000", b.WebhookHandler())
	if err != nil {
		log.Fatal(err)
	}

	// call methods.DeleteWebhook if needed

}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("Got a message %s, and will resend it back.", update.Message.Text)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
	if err != nil {
		log.Print(err)
	}
}
