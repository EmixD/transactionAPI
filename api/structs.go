package main

import "time"

type User struct {
	Id           uint64  `json:"id" bson:"_id"`
	Balance      float64 `json:"balance"`
	DepositCount uint64  `json:"depositcount"`
	DepositSum   float64 `json:"depositsum"`
	BetCount     uint64  `json:"betcount"`
	BetSum       float64 `json:"betsum"`
	WinCount     uint64  `json:"wincount"`
	WinSum       float64 `json:"winsum"`
}

type Deposit struct {
	DepositId     uint64    `json:"depositid" bson:"_id"`
	UserId        uint64    `json:"userid"`
	Amount        float64   `json:"amount"`
	BalanceBefore float64   `json:"balancabefore"`
	BalanceAfter  float64   `json:"balanceafter"`
	Time          time.Time `json:"time"`
}

type Transaction struct {
	TransactionId uint64    `json:"transactionid" bson:"_id"`
	UserId        uint64    `json:"userid"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	BalanceBefore float64   `json:"balancebefore"`
	BalanceAfter  float64   `json:"balanceafter"`
	Time          time.Time `json:"time"`
}

type AddUserInput struct {
	Id      uint64  `json:"id" binding:"required"`
	Balance float64 `json:"balance" binding:"required"`
	Token   string  `json:"token" binding:"required"`
}

type GetUserInput struct {
	Id    uint64 `json:"id" binding:"required"`
	Token string `json:"token" binding:"required"`
}

type AddDepositInput struct {
	DepositId uint64  `json:"depositid" binding:"required"`
	UserId    uint64  `json:"userid" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Token     string  `json:"token" binding:"required"`
}

type AddTransactionInput struct {
	TransactionId uint64  `json:"transactionid"  binding:"required"`
	UserId        uint64  `json:"userid" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	Token         string  `json:"token" binding:"required"`
}
