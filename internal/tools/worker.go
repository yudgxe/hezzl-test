package tools

import (
	"github.com/rs/zerolog/log"

	"github.com/nats-io/nats.go"
)

type Sender interface {
	Send(data interface{}) error
}

type Worker[T any] struct {
	sender    Sender
	nats      *nats.EncodedConn
	buff      []T
	batchSize int

	onclose func()
}

func NewWorker[T any](nats *nats.EncodedConn, sender Sender, batchSize int) *Worker[T] {
	return &Worker[T]{
		nats:      nats,
		batchSize: batchSize,
		buff:      make([]T, 0, batchSize),
		sender:    sender,
	}
}

func (w *Worker[T]) Start(subject string) error {
	sub, err := w.nats.Subscribe(subject, func(in T) {
		w.buff = append(w.buff, in)
		if len(w.buff) == w.batchSize {
			if err := w.sender.Send(w.buff); err != nil {
				log.Error().Err(err).Msg("failed to send")
			}
			w.buff = make([]T, 0, w.batchSize)
		}
	})
	if err != nil {
		return err
	}
	w.onclose = func() {
		if err := sub.Unsubscribe(); err != nil {
			log.Error().Err(err).Msg("failed to unsubscribe")
		}
	}
	return nil
}

func (w *Worker[T]) Stop() {
	w.onclose()
	w.sender.Send(w.buff)
}
