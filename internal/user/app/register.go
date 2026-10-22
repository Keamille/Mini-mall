package app

import (
	"database/sql"
	"github.com/your-org/go-minimall/internal/user/domain"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Nickname   string `json:"nickname" binding:"required"`
	Phone      string `json:"phone"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
	Email      string `json:"email" binding:"email"`
}

func (h *Handler) Register_user(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
		return
	}
	if req.Phone == "" && req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "phone or email is required"})
		return
	}

	id := domain.NewID()

	var email, phone any
	if req.Email != "" {
		email = req.Email
	}
	if req.Phone != "" {
		phone = req.Phone
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate password hash", "error": err.Error()})
		return
	}

	if err := h.Repo.CreateUser(id, email, phone, string(hash), req.Nickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to register user", "error": err.Error()})
		return
	}

	h.respondWithTokens(c, http.StatusCreated, domain.UserRow{
		ID:           id,
		Email:        sql.NullString{String: req.Email, Valid: req.Email != ""},
		Phone:        sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		PasswordHash: string(hash),
		Nickname:     req.Nickname,
		Status:       1,
	})
}
