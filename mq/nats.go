package mq

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func GetNats() *nats.Conn {
	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		err = fmt.Errorf("failed to connect to nats with err: %w", err)
		log.Fatalln(err.Error())
	}
	return nc
}

func GetJetStream(nc *nats.Conn) jetstream.JetStream {
	js, err := jetstream.New(nc)
	if err != nil {
		err = fmt.Errorf("failed to get jetstream with err: %w", err)
		log.Fatalln(err.Error())
	}
	return js
}
