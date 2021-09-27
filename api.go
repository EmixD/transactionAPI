package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var UserRefs = map[uint64]*User{}                         // All users
var DepositRefs = map[uint64]*Deposit{}                   // All deposits
var TransactionRefs = map[uint64]*Transaction{}           // All transactions
var UserRefsNeedUpdate = map[uint64]*User{}               // Users that need to be updated in DB
var DepositRefsNeedUpdate = map[uint64]*Deposit{}         // Deposits that need to be updated in DB
var TransactionRefsNeedUpdate = map[uint64]*Transaction{} // Transactions that need to be updated in DB

func AddUser(c *gin.Context) {
	var input AddUserInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, isInUserRefs := UserRefs[input.Id]
	if isInUserRefs {
		c.JSON(http.StatusConflict, gin.H{"error": "A player with this ID already exists"})
		return
	}

	if input.Token != "testtask" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		return
	}

	newUser := new(User)
	newUser.Id = input.Id
	newUser.Balance = input.Balance
	UserRefs[newUser.Id] = newUser
	UserRefsNeedUpdate[newUser.Id] = newUser
	c.IndentedJSON(http.StatusCreated, gin.H{"error": ""})
}

func GetUser(c *gin.Context) {
	var input GetUserInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, isInUserRefs := UserRefs[input.Id]
	if !isInUserRefs {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if input.Token != "testtask" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		return
	}
	c.IndentedJSON(http.StatusOK, UserRefs[input.Id])
}

func AddDeposit(c *gin.Context) {
	var input AddDepositInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, isInUserRefs := UserRefs[input.UserId]
	if !isInUserRefs {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	_, isInDepositRefs := DepositRefs[input.DepositId]
	if isInDepositRefs {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A deposit with this ID already exists"})
		return
	}

	if input.Token != "testtask" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		return
	}

	newDeposit := new(Deposit)
	newDeposit.DepositId = input.DepositId
	newDeposit.UserId = input.UserId
	newDeposit.Amount = input.Amount
	newDeposit.BalanceBefore = UserRefs[input.UserId].Balance
	newDeposit.BalanceAfter = UserRefs[input.UserId].Balance + input.Amount
	newDeposit.Time = time.Now()
	DepositRefs[input.DepositId] = newDeposit
	DepositRefsNeedUpdate[input.DepositId] = newDeposit

	UserRefs[input.UserId].Balance += input.Amount
	UserRefs[input.UserId].DepositSum += input.Amount
	UserRefs[input.UserId].DepositCount++
	UserRefsNeedUpdate[input.UserId] = UserRefs[input.UserId]

	c.JSON(http.StatusCreated, gin.H{"error": "", "balance": UserRefs[input.UserId].Balance})
}

func AddTransaction(c *gin.Context) {
	var input AddTransactionInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, isInUserRefs := UserRefs[input.UserId]
	if !isInUserRefs {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	_, isInTransactionRefs := TransactionRefs[input.TransactionId]
	if isInTransactionRefs {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A transaction with this ID already exists"})
		return
	}

	if input.Token != "testtask" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token"})
		return
	}

	balanceBefore := UserRefs[input.UserId].Balance

	switch input.Type {
	case "Win":
		UserRefs[input.UserId].Balance += input.Amount
		UserRefs[input.UserId].WinSum += input.Amount
		UserRefs[input.UserId].WinCount++
	case "Bet":
		if UserRefs[input.UserId].Balance < input.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient user balance"})
			return
		}
		UserRefs[input.UserId].Balance -= input.Amount
		UserRefs[input.UserId].BetSum += input.Amount
		UserRefs[input.UserId].BetCount++
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect transaction type"})
		return
	}

	newTransaction := new(Transaction)
	newTransaction.TransactionId = input.TransactionId
	newTransaction.UserId = input.UserId
	newTransaction.Type = input.Type
	newTransaction.Amount = input.Amount
	newTransaction.BalanceBefore = balanceBefore
	newTransaction.BalanceAfter = UserRefs[input.UserId].Balance
	newTransaction.Time = time.Now()

	TransactionRefs[input.TransactionId] = newTransaction
	TransactionRefsNeedUpdate[input.TransactionId] = newTransaction

	UserRefsNeedUpdate[input.UserId] = UserRefs[input.UserId]
	c.JSON(http.StatusCreated, gin.H{"error": "", "balance": UserRefs[input.UserId].Balance})
}
