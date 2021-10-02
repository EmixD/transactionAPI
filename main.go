package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
)

func StartServer() *http.Server {
	router := gin.Default()
	router.POST("/user/create", AddUser)
	router.POST("/user/get", GetUser)
	router.POST("/user/deposit", AddDeposit)
	router.POST("/transaction", AddTransaction)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println("Http listener stopped: ", err)
		}
	}()
	return srv
}

func main() {
	dbUpdatePeriod := time.Second * 4
	dbUpdateMaxSyncTime := time.Second * 2
	chStopLoop := make(chan int) // Any data sent to this chan will stop sync with DB

	DbConnect()
	go DbSyncLoop(chStopLoop, dbUpdatePeriod, dbUpdateMaxSyncTime)

	srv := StartServer()

	quit := make(chan os.Signal, 100)
	signal.Notify(quit, os.Interrupt)

	<-quit
	fmt.Println("Shutting down the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Stopping sync loop...")
	chStopLoop <- 1

	fmt.Println("Performing last sync...")
	DbUpdate(dbUpdateMaxSyncTime)

	fmt.Println("Disconnecting from DB...")
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
	DbClient.Disconnect(ctx)
	ctxCancel()
	DbCtxConnectCancel()

	fmt.Println("SHUTDOWN COMPLETE")
}
