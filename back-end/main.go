package main

import (
	"chat-app/back-end/application"
	"context"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	app := application.New()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("Failed to start app: ", err)
	}
}
