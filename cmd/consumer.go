package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"nats_tracing/mq"
	inOtel "nats_tracing/otel"
)

func ConsumerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "consumer",
		Short:   "consumer of message",
		Aliases: []string{"c"},
		Example: "nats_tracing consumer",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.ValidateArgs(args); err != nil {
				err = fmt.Errorf("argument validation failed: %w", err)
				log.Fatalln(err.Error())
			}
			startConsumer(cmd.Context())
		},
	}
	return cmd
}

func startConsumer(ctx context.Context) {
	shutdowns, err := inOtel.InitOtelSdk(ctx, "consumer")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer inOtel.ShutdownOtel(ctx, shutdowns)

	nc := mq.GetNats()
	defer nc.Close()

	js := mq.GetJetStream(nc)

	st, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "stream:download",
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{"events.>"},
	})
	if err != nil {
		err = fmt.Errorf("failed to create or update stream with err: %w", err)
		log.Fatalln(err.Error())
	}

	cs, err := st.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{})
	if err != nil {
		err = fmt.Errorf("failed to create consumer with err: %w", err)
		log.Fatalln(err.Error())
	}
	cc, err := cs.Consume(func(msg jetstream.Msg) {
		header := msg.Headers()
		carrier := propagation.HeaderCarrier(header)
		ctx := inOtel.GetTextMapPropagator().Extract(context.Background(), carrier)

		ctx, span := otel.Tracer("consumer").Start(ctx, "received message")
		defer span.End()

		handleMessage(ctx, msg)

		defer func() {
			log.Println("acknowledging message")
			if err := msg.Ack(); err != nil {
				log.Printf("failed to acknowledge message with err: %s", err.Error())
				return
			}
			log.Println("message acknowledged")
		}()
	})
	if err != nil {
		err = fmt.Errorf("failed to consume with err: %w", err)
		log.Fatalln(err.Error())
	}
	defer func() {
		log.Println("stopping consumer connection")
		cc.Stop()
	}()

	log.Println("consumer started")
	<-ctx.Done()
}

func handleMessage(ctx context.Context, msg jetstream.Msg) {
	ctx, span := otel.Tracer("consumer").Start(ctx, "handleMessage")
	defer span.End()

	log.Printf("message received from subject: %s, message data: %s", msg.Subject(), msg.Data())
}
