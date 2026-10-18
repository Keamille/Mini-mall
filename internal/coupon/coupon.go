package coupon

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Coupon struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          int64     `json:"type"`
	ThresholdCent int64     `json:"threshold_cent"`
	DiscountCent  int64     `json:"discount_cent"`
	DiscountRate  int64     `json:"discount_rate"`
	TotalCount    int64     `json:"total_count"`
	ClaimedCount  int64     `json:"claimed_count"`
	PerUserLimit  int64     `json:"per_user_limit"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidTo       time.Time `json:"valid_to"`
	Status        int       `json:"status"`
}

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.listTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list coupons"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": templates})
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var rep Coupon
	if err := c.ShouldBindJSON(&rep); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request body"})
		return
	}
	rep.ID = NewID()
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
	if err := h.createTemplate(rep); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create template"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "success", "data": rep})
}

func (h *Handler) Claim(c *gin.Context) {
	userID := c.GetString("user_id")
	templateID := c.Param("id")
	row, err := h.claimTx(userID, templateID)
	if errors.Is(err, ErrTemplateNotClaimable) {
		c.JSON(http.StatusNotFound, gin.H{"code": 17001, "message": "template not claimable"})
		return
	}
	if errors.Is(err, ErrClaimLimit) {
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
	coupons, err := h.listMine(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list mine"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": coupons})
}
