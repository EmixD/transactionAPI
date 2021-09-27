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
	dbUpdatePeriod := time.Second * 10
	dbUpdateMaxSyncTime := time.Second * 5
	DbStartSync(dbUpdatePeriod, dbUpdateMaxSyncTime)

	srv := StartServer()
	quit := make(chan os.Signal, 100)
	signal.Notify(quit, os.Interrupt)

	<-quit
	fmt.Println("Stopping sync...")
	DbStopSync(dbUpdatePeriod + dbUpdateMaxSyncTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Shutting down...")
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("SHUTDOWN COMPLETE")
}
