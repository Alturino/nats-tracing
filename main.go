package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"nats_tracing/cmd"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	rootCmd := &cobra.Command{}
	rootCmd.AddCommand(cmd.ConsumerCmd(), cmd.Publisher())
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatalln(err.Error())
	}
}
