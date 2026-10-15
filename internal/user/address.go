package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddressRequest struct {
	ReceiverName string `json:"receiver_name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Province     string `json:"province"`
	City         string `json:"city"`
	District     string `json:"district"`
	Detail       string `json:"detail" binding:"required"`
	IsDefault    *bool  `json:"is_default"` // 指针：区分「没传」和 false
}

func (h *Handler) ListAddress(c *gin.Context) {
	userID := c.GetString("user_id")
	rows, err := h.listAddresses(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to list addresses"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    rows,
	})
}

func (h *Handler) CreateAddress(c *gin.Context) {
	userID := c.GetString("user_id")
	var req AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}

	address := addressRow{
		ID:           NewID(),
		UserID:       userID,
		ReceiverName: req.ReceiverName,
		Phone:        req.Phone,
		Province:     req.Province,
		City:         req.City,
		District:     req.District,
		Detail:       req.Detail,
	}
	if req.IsDefault != nil && *req.IsDefault {
		address.IsDefault = 1
	}

	if err := h.createAddressTx(address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to create address"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "success",
		"data":    address,
	})
}

func (h *Handler) UpdateAddress(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var req AddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 10001, "message": "invalid request"})
		return
	}

	address, err := h.findAddressByID(id)
	if err != nil || address.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"code": 11005, "message": "address not found"})
		return
	}

	address.ReceiverName = req.ReceiverName
	address.Phone = req.Phone
	address.Province = req.Province
	address.City = req.City
	address.District = req.District
	address.Detail = req.Detail

	setDefault := false
	if req.IsDefault != nil {
		setDefault = *req.IsDefault
		if *req.IsDefault {
			address.IsDefault = 1
		} else {
			address.IsDefault = 0
		}
	}

	if err := h.updateAddressTx(address, setDefault); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to update address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    address,
	})
}

func (h *Handler) DeleteAddress(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	address, err := h.findAddressByID(id)
	if err != nil || address.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"code": 11005, "message": "address not found"})
		return
	}

	if err := h.deleteAddress(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 10007, "message": "failed to delete address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}
