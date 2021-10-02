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
var DbCtxConnectCancel context.CancelFunc // Cancel function for the dbClient.Connect context
var DbClient *mongo.Client

func DbConnect() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal(err)
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

func DbSyncLoop(chStopLoop chan int, period time.Duration, maxsynctime time.Duration) {
	TimeToSync := time.After(period)
	for {
		select {
		case <-TimeToSync:
			DbUpdate(maxsynctime)
			TimeToSync = time.After(period)
		case <-chStopLoop:
			fmt.Println("DB sync loop stopped...")
			return
		}
	}
}

func DbUpdate(maxtime time.Duration) {
	mutex.Lock()
	UserRefsNeedUpdateCopy := make([]*User, len(UserRefsNeedUpdate))
	DepositRefsNeedUpdateCopy := make([]*Deposit, len(DepositRefsNeedUpdate))
	TransactionRefsNeedUpdateCopy := make([]*Transaction, len(TransactionRefsNeedUpdate))
	copy(UserRefsNeedUpdateCopy, UserRefsNeedUpdate)
	copy(DepositRefsNeedUpdateCopy, DepositRefsNeedUpdate)
	copy(TransactionRefsNeedUpdateCopy, TransactionRefsNeedUpdate)
	UserRefsNeedUpdate = []*User{}
	DepositRefsNeedUpdate = []*Deposit{}
	TransactionRefsNeedUpdate = []*Transaction{}
	mutex.Unlock()

	// If any of the User, Deposit or Transaction objects gets modified while this goroutine executes,
	// any possible error will be corrected on the next call

	ctx, cancel := context.WithTimeout(context.Background(), maxtime)
	defer cancel()

	for _, u := range UserRefsNeedUpdateCopy {
		_, err := ColUsers.ReplaceOne(ctx,
			bson.D{{Key: "_id", Value: u.Id}},
			u,
			options.Replace().SetUpsert(true))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	for _, d := range DepositRefsNeedUpdateCopy {
		_, err := ColDeposits.InsertOne(ctx, d)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	for _, d := range TransactionRefsNeedUpdateCopy {
		_, err := ColTransactions.InsertOne(ctx, d)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
