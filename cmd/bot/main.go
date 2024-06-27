package main

import (
	"context"
	telegramBot "github.com/go-telegram/bot"
	"log"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go startHealthCheckServer(ctx)

	b := newBot()

	opts := []telegramBot.Option{
		telegramBot.WithDefaultHandler(b.defaultHandler),
	}

	tgBot, err := telegramBot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)
	if err != nil {
		log.Fatal(err)
	}

	tgBot.RegisterHandler(
		telegramBot.HandlerTypeMessageText,
		"/start",
		telegramBot.MatchTypeExact,
		b.startHandler,
	)

	tgBot.RegisterHandlerMatchFunc(b.matchGetResults, b.getResultsHandler)

	log.Print("Starting the bot server...")

	tgBot.Start(ctx)

}
