package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sw "github.com/KubeQuest902/project/projectserver"
	"github.com/mediocregopher/radix/v3"
)

const (
	defaultExposePort = "8080"
)

func main() {
	var ok bool
	var err error

	sw.RedisHost, ok = os.LookupEnv("REDIS_HOST")
	if !ok {
		log.Fatal("Error: could not read $REDIS_HOST")
	}
	sw.RedisPort, ok = os.LookupEnv("REDIS_PORT")
	if !ok {
		log.Fatal("Error: could not read $REDIS_PORT")
	}
	sw.RedisPassword, ok = os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		log.Fatal("Error: could not read $REDIS_PASSWORD")
	}
	sw.RedisPool, err = radix.NewPool(
		"tcp",
		sw.RedisHost+":"+sw.RedisPort,
		10,
		radix.PoolConnFunc(func(network, addr string) (radix.Conn, error) {
			return radix.Dial(network, addr,
				radix.DialAuthPass(sw.RedisPassword),
			)
		}),
	)
	if err != nil {
		log.Fatalf("failed to create redis pool: %v", err)
	}

	exposePort, ok := os.LookupEnv("EXPOSE_PORT")
	if !ok {
		exposePort = defaultExposePort
	}
	log.Printf(fmt.Sprintf("Server listens on port %v", exposePort))

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGTERM) // Received after the preStop hook

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", 8080),
		Handler: sw.NewRouter(),
	}

	go server.ListenAndServe()
	go sw.StartWebSocketBroadcaster()

	select {
	case c := <-termChan:
		log.Printf("Received signal %v, stopping gracefully", c)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		server.Shutdown(ctx)
		log.Printf("Server stopped, exiting. ")
	}
}
