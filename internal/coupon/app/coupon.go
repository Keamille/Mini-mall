package app

import (
	"errors"
	"net/http"

	"github.com/your-org/go-minimall/internal/coupon/domain"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.Repo.ListTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list coupons"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": templates})
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var rep domain.Coupon
	if err := c.ShouldBindJSON(&rep); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request body"})
		return
	}
	rep.ID = domain.NewID()
	rep.Type = 1
	rep.Status = 1
	rep.ClaimedCount = 0
	rep.DiscountRate = 0
	if rep.PerUserLimit == 0 {
		rep.PerUserLimit = 1
	}
	if rep.Name == "" || rep.DiscountCent <= 0 || rep.TotalCount <= 0 || rep.ThresholdCent < 0 || rep.PerUserLimit < 1 || !rep.ValidTo.After(rep.ValidFrom) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request body"})
		return
	}
	if err := h.Repo.CreateTemplate(rep); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create template"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "success", "data": rep})
}

func (h *Handler) Claim(c *gin.Context) {
	userID := c.GetString("user_id")
	templateID := c.Param("id")
	row, err := h.Repo.ClaimTx(userID, templateID)
	if errors.Is(err, domain.ErrTemplateNotClaimable) {
		c.JSON(http.StatusNotFound, gin.H{"code": 17001, "message": "template not claimable"})
		return
	}
	if errors.Is(err, domain.ErrClaimLimit) {
		c.JSON(http.StatusConflict, gin.H{"code": 17002, "message": "claim limit reached"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to claim"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "success", "data": row})
}

func (h *Handler) ListMine(c *gin.Context) {
	userID := c.GetString("user_id")
	coupons, err := h.Repo.ListMine(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list mine"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": coupons})
}
