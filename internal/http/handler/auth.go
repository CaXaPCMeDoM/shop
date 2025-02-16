package handler

import (
	"avito-tech-winter-2025/internal/config"
	"avito-tech-winter-2025/internal/http/auth"
	"avito-tech-winter-2025/internal/storage"
	"avito-tech-winter-2025/pkg/hash"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo storage.Storage
	tokenMgr *auth.Manager
	hasher   *hash.SHA1
	cfg      *config.Config
}

func NewAuthHandler(repo storage.Storage, mgr *auth.Manager, hasher *hash.SHA1, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo: repo,
		tokenMgr: mgr,
		hasher:   hasher,
		cfg:      cfg,
	}
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		var hashErr error
		hashedPassword, hashErr := h.hasher.Hash(req.Password)
		if hashErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		user, err = h.userRepo.Create(req.Username, hashedPassword)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
	} else {
		if compareErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); compareErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
	}

	accessToken, err := h.tokenMgr.NewJwt(user.ID, h.cfg.Jwt.TokenTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: accessToken,
	})
}
