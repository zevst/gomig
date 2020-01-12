package util

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func RegisterCloser() context.Context {
	closer := make(chan os.Signal, 1)
	signal.Notify(closer, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go handler(closer, cancel)
	return ctx
}

func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Println(err)
	}
}

func handler(close <-chan os.Signal, cancel context.CancelFunc) {
	<-close
	cancel()
}
