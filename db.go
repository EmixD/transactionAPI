package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ColUsers *mongo.Collection
var ColDeposits *mongo.Collection
var ColTransactions *mongo.Collection

var dbCtxSyncCancel context.CancelFunc
var dbCtxConnectCancel context.CancelFunc
var dbClient *mongo.Client

func dbConnect() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbClient, err = mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URL")))
	if err != nil {
		log.Fatal(err)
	}
	var dbCtxConnect context.Context
	dbCtxConnect, dbCtxConnectCancel = context.WithCancel(context.Background())

	err = dbClient.Connect(dbCtxConnect)
	if err != nil {
		log.Fatal(err)
	}

	ColUsers = dbClient.Database("dbname").Collection("colUsers")
	ColDeposits = dbClient.Database("dbname").Collection("colDeposits")
	ColTransactions = dbClient.Database("dbname").Collection("colTransactions")

	err = ColUsers.Drop(dbCtxConnect)
	if err != nil {
		log.Fatal(err)
	}
	err = ColDeposits.Drop(dbCtxConnect)
	if err != nil {
		log.Fatal(err)
	}
	err = ColTransactions.Drop(dbCtxConnect)
	if err != nil {
		log.Fatal(err)
	}
}

func dbStartSync(period time.Duration, maxsynctime time.Duration) {
	dbConnect()
	var ctxSync context.Context
	ctxSync, dbCtxSyncCancel = context.WithCancel(context.Background())
	go dbSyncLoop(ctxSync, period, maxsynctime)
}

func dbStopSync(maxsynctime time.Duration) {
	dbCtxSyncCancel()
	time.Sleep(maxsynctime) // Wait for current sync if any
	dbUpdate(maxsynctime)   // last DB update
	ctx, ctxCancel := context.WithTimeout(context.Background(), maxsynctime)
	dbClient.Disconnect(ctx)
	ctxCancel()
	dbCtxConnectCancel()
}

func dbSyncLoop(ctxsync context.Context, period time.Duration, maxsynctime time.Duration) {
	for i := 0; i < 20; i++ { //should be replaced by an infinite loop
		time.Sleep(period)
		if ctxsync.Err() != nil {
			fmt.Println("Sync loop stopped")
			return
		}
		fmt.Println("Starting sync...")
		dbUpdate(maxsynctime)
		fmt.Println("Sync done")
	}
}

func dbUpdate(maxtime time.Duration) {
	UserRefsNeedUpdateCopy := UserRefsNeedUpdate
	DepositRefsNeedUpdateCopy := DepositRefsNeedUpdate
	TransactionRefsNeedUpdateCopy := TransactionRefsNeedUpdate

	UserRefsNeedUpdate = map[uint64]*User{}
	DepositRefsNeedUpdate = map[uint64]*Deposit{}
	TransactionRefsNeedUpdate = map[uint64]*Transaction{}

	ctx, cancel := context.WithTimeout(context.Background(), maxtime)
	defer cancel()

	for _, u := range UserRefsNeedUpdateCopy {
		fmt.Println("Now upserting user: ", u)
		_, err := ColUsers.ReplaceOne(ctx,
			bson.D{{Key: "_id", Value: u.Id}},
			u,
			options.Replace().SetUpsert(true))
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, d := range DepositRefsNeedUpdateCopy {
		fmt.Println("Now inserting deposit: ", d)
		_, err := ColDeposits.InsertOne(ctx, d)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, d := range TransactionRefsNeedUpdateCopy {
		fmt.Println("Now inserting transaction: ", d)
		_, err := ColTransactions.InsertOne(ctx, d)
		if err != nil {
			log.Fatal(err)
		}
	}

}
