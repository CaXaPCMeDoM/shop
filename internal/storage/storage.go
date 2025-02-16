package storage

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidItem       = errors.New("invalid item")
)

type Storage interface {
	GetByUsername(username string) (*User, error)
	Create(username, passwordHash string) (*User, error)
	TransferCoins(ctx context.Context, fromUserID int, toUsername string, amount int) error
	GetUserInfo(ctx context.Context, userID int) (*UserInfo, error)
	BuyItem(ctx context.Context, userID int, item string) error
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Coins        int
}

type UserInfo struct {
	Coins       int
	Inventory   []InventoryItem
	CoinHistory CoinHistory
}

type InventoryItem struct {
	Type     string
	Quantity int
}

type CoinHistory struct {
	Received []Transaction
	Sent     []Transaction
}

type Transaction struct {
	FromUser  string
	ToUser    string
	Amount    int
	CreatedAt string
}
