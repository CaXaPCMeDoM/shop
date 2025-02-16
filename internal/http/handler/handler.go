package handler

import (
	"avito-tech-winter-2025/internal/storage"
)

type Handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {
	return &Handler{storage: storage}
}
