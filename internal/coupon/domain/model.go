package domain

import (
	"errors"
	"time"
)

type CouponRow struct {
	ID            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Type          int64     `db:"type" json:"type"`
	ThresholdCent int64     `db:"threshold_cent" json:"threshold_cent"`
	DiscountCent  int64     `db:"discount_cent" json:"discount_cent"`
	DiscountRate  int64     `db:"discount_rate" json:"discount_rate"`
	TotalCount    int64     `db:"total_count" json:"total_count"`
	ClaimedCount  int64     `db:"claimed_count" json:"claimed_count"`
	PerUserLimit  int64     `db:"per_user_limit" json:"per_user_limit"`
	ValidFrom     time.Time `db:"valid_from" json:"valid_from"`
	ValidTo       time.Time `db:"valid_to" json:"valid_to"`
	Status        int       `db:"status" json:"status"`
}

type UserCouponRow struct {
	ID         string `db:"id" json:"id"`
	TemplateID string `db:"template_id" json:"template_id"`
	UserID     string `db:"user_id" json:"user_id"`
	Status     int    `db:"status" json:"status"`
}

type MineRow struct {
	ID            string    `db:"id" json:"id"`
	TemplateID    string    `db:"template_id" json:"template_id"`
	Status        int       `db:"status" json:"status"`
	LockedOrderID string    `db:"locked_order_id" json:"locked_order_id"`
	Name          string    `db:"name" json:"name"`
	ThresholdCent int64     `db:"threshold_cent" json:"threshold_cent"`
	DiscountCent  int64     `db:"discount_cent" json:"discount_cent"`
	ValidTo       time.Time `db:"valid_to" json:"valid_to"`
}

type CouponState struct {
	ID            string `db:"id"`
	UserID        string `db:"user_id"`
	Status        int    `db:"status"`
	LockedOrderID string `db:"locked_order_id"`
	ThresholdCent int64  `db:"threshold_cent"`
	DiscountCent  int64  `db:"discount_cent"`
	Expired       int    `db:"expired"`
}

type CouponEffect struct {
	UserCouponID string `json:"user_coupon_id"`
	DiscountCent int64  `json:"discount_cent"`
	Status       int    `json:"status"`
}

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

var (
	ErrTemplateNotClaimable = errors.New("template not claimable")
	ErrClaimLimit           = errors.New("claim limit reached")
	ErrCouponNotFound       = errors.New("coupon not found")
	ErrCouponStatus         = errors.New("coupon status invalid")
	ErrCouponThreshold      = errors.New("threshold not met")
	ErrCouponExpired        = errors.New("coupon expired")
)
