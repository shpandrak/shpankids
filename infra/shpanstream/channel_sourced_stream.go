package shpanstream

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"
)

const defaultMaxAllowedChannelDelay = time.Second
const defaultBufferSize = 50

type channelProvider[T any] struct {
	originalChannel       <-chan T
	intermediateChannel   chan T
	maxAllowedReaderDelay time.Duration
	bufferSize            int
	readerSlowCtx         context.Context
	readerSlowCtxCancel   context.CancelFunc
}

func (cp *channelProvider[T]) Open(_ context.Context) error {
	// Opening, this means we are starting to read from the channel
	// We will start a goroutine that will read from the original channel and write to the intermediate channel
	//respecting the max delay
	cp.intermediateChannel = make(chan T, cp.bufferSize)
	cp.readerSlowCtx, cp.readerSlowCtxCancel = context.WithCancel(context.Background())

	go func() {
		defer close(cp.intermediateChannel)
		for {
			select {
			case msg, stillGood := <-cp.originalChannel:
				if !stillGood {
					slog.Debug("original channel closed, closing the intermediate channel")
					cp.readerSlowCtxCancel()
				}
				select {
				case cp.intermediateChannel <- msg:
				case <-time.After(cp.maxAllowedReaderDelay):
					cp.readerSlowCtxCancel()
					slog.Debug(fmt.Sprintf(
						"stream closed due to max delay while writing to intermediate channel %s",
						cp.maxAllowedReaderDelay,
					))
					return
				}
			}
		}
	}()
	return nil
}

func (cp *channelProvider[T]) Close() {
	slog.Debug("closing channel stream provider")
}

func (cp *channelProvider[T]) Emit(ctx context.Context) (*T, error) {
	select {
	case <-ctx.Done():
		slog.Debug("stream channel ctx closed, finishing the stream")
		return nil, io.EOF
	case <-cp.readerSlowCtx.Done():
		return nil, fmt.Errorf("stream context closed due to reader being too slow %s", cp.maxAllowedReaderDelay)
	case msg, stillGood := <-cp.intermediateChannel:
		if !stillGood {
			slog.Debug("stream channel closed externally")
			return nil, fmt.Errorf("stream closed due to reader being too slow %s", cp.maxAllowedReaderDelay)
		}
		return &msg, nil
	}
}

func FromChannel[T any](ch <-chan T) Stream[T] {
	return FromChannelWithOpts(ch, defaultMaxAllowedChannelDelay, defaultBufferSize)
}

func FromChannelWithOpts[T any](ch <-chan T, maxAllowedReaderDelay time.Duration, bufferSize int) Stream[T] {
	return NewStream[T](&channelProvider[T]{
		originalChannel:       ch,
		maxAllowedReaderDelay: maxAllowedReaderDelay,
		bufferSize:            bufferSize,
	})
}
