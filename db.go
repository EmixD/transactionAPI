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

var DbCtxSyncCancel context.CancelFunc    // Function to be called to signal a need to stop sync with DB
var DbCtxConnectCancel context.CancelFunc // Cancel function for the dbClient.Connect context
var DbClient *mongo.Client

func DbConnect() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	DbClient, err = mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URL")))
	if err != nil {
		log.Fatal(err)
	}

	var DbCtxConnect context.Context // This context will be active while the server runs
	DbCtxConnect, DbCtxConnectCancel = context.WithCancel(context.Background())

	err = DbClient.Connect(DbCtxConnect)
	if err != nil {
		log.Fatal(err)
	}

	ColUsers = DbClient.Database(os.Getenv("DBNAME")).Collection(os.Getenv("COLLECTION_USERS_NAME"))
	ColDeposits = DbClient.Database(os.Getenv("DBNAME")).Collection(os.Getenv("COLLECTION_DEPOSITS_NAME"))
	ColTransactions = DbClient.Database(os.Getenv("DBNAME")).Collection(os.Getenv("COLLECTION_TRANSACTIONS_NAME"))

	// err = ColUsers.Drop(dbCtxConnect)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = ColDeposits.Drop(dbCtxConnect)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = ColTransactions.Drop(dbCtxConnect)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func DbStartSync(period time.Duration, maxsynctime time.Duration) {
	// Start periodic sync with DB
	DbConnect()

	var ctxSync context.Context
	ctxSync, DbCtxSyncCancel = context.WithCancel(context.Background())
	// dbCtxSyncCancel is used later to terminate dbSyncLoop
	go DbSyncLoop(ctxSync, period, maxsynctime)
}

func DbStopSync(waittime time.Duration) {
	// Stop periodic sync with DB
	DbCtxSyncCancel()
	fmt.Println("Waiting for DB sync loop to stop...")
	time.Sleep(waittime) // Wait for current sync if any

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second)
	DbClient.Disconnect(ctx)
	ctxCancel()

	DbCtxConnectCancel()
}

func DbSyncLoop(ctxsync context.Context, period time.Duration, maxsynctime time.Duration) {
	for {
		time.Sleep(period)
		DbUpdate(maxsynctime)
		if ctxsync.Err() != nil {
			fmt.Println("DB sync loop stopped")
			return
		}
	}
}

func DbUpdate(maxtime time.Duration) {
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
