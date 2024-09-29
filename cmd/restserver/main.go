package main

import (
	"context"
	"fmt"
	"gfsloader/cmd/restserver/handlers"
	"gfsloader/cmd/restserver/serverapp"
	"gfsloader/internal/storage/postgres"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	apiBasePath = "api/v1"
)

func main() {

	ctx := context.TODO()

	storageProvider := postgres.New("host=localhost port=5555 user=postgres dbname=weather password=postgres sslmode=disable")
	storageProvider.MustRun()

	wktHandler := handlers.NewWKTHandler(storageProvider)

	serverApp := serverapp.New(apiBasePath, wktHandler)

	errSig := make(chan error)
	stopSig := make(chan os.Signal, 1)

	go func() {
		err := serverApp.Run("localhost", 8080)
		if err != nil {
			errSig <- err
		}
	}()

	signal.Notify(stopSig, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-errSig:
		panic(err)
	case <-stopSig:
		waitCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		err := serverApp.Stop(waitCtx)
		if err != nil {
			panic(err)
		}
		cancel()
		fmt.Println("Server stopped")
	}

}
