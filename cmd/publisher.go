package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/cobra"
	"github.com/xyproto/randomstring"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"nats_tracing/mq"
	inOtel "nats_tracing/otel"
)

func Publisher() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "publisher",
		Short:   "publisher message",
		Aliases: []string{"p"},
		Example: "nats_tracing publisher",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.ValidateArgs(args); err != nil {
				err = fmt.Errorf("argument validation failed: %w", err)
				log.Fatalln(err.Error())
			}
			startPublisher(cmd.Context())
		},
	}
	return cmd
}

func startPublisher(ctx context.Context) {
	shutdowns, err := inOtel.InitOtelSdk(ctx, "publisher")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer inOtel.ShutdownOtel(ctx, shutdowns)

	nc := mq.GetNats()
	defer func() {
		log.Println("closing nats connection")
		nc.Close()
		log.Println("closed nats connection")
	}()

	js := mq.GetJetStream(nc)

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "stream:download",
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{"events.>"},
	})
	if err != nil {
		err = fmt.Errorf("failed to create or update stream with err: %w", err)
		log.Fatalln(err.Error())
	}

	go func(ctx context.Context) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		log.Println("publisher started")
		for range time.Tick(time.Second * 5) {
			timeToPublish(ctx, js)
		}
	}(ctx)

	<-ctx.Done()
}

func timeToPublish(ctx context.Context, js jetstream.JetStream) {
	ctx, span := otel.Tracer("publisher").Start(ctx, "timeToPublish")
	defer span.End()

	publish(ctx, js)
}

func publish(ctx context.Context, js jetstream.JetStream) {
	ctx, span := otel.Tracer("publisher").Start(ctx, "publish")
	defer span.End()

	header := make(nats.Header)
	carrier := propagation.HeaderCarrier(header)
	inOtel.GetTextMapPropagator().Inject(ctx, carrier)

	log.Println("publisher headers:", header)

	msgID := randomstring.HumanFriendlyEnglishString(16)
	fut, err := js.PublishMsgAsync(
		&nats.Msg{
			Header:  header,
			Data:    []byte("hello from sender"),
			Subject: "events.download." + msgID,
		},
		jetstream.WithMsgID(msgID),
	)
	if err != nil {
		err = fmt.Errorf("failed to PublishMsgAsync with err: %w", err)
		log.Fatalln(err.Error())
	}
	log.Printf("published message to subject: %s", fut.Msg().Subject)
}
