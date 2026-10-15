package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Phone    string `json:"phone"`
	Email    string `json:"email" binding:"email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}
	if req.Email == "" && req.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "email or phone required"})
		return
	}

	var u userRow
	var err error
	if req.Email != "" {
		u, err = h.findUserByEmail(req.Email)
	} else {
		u, err = h.findUserByPhone(req.Phone)
	}
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 11002, "message": "invalid account or password"})
		return
	}
	if u.Status != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 11004, "message": "user disabled"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 11002, "message": "invalid account or password"})
		return
	}

	h.respondWithTokens(c, http.StatusOK, u)
}
