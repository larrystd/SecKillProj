package model

type Coupon struct {
	Id          int64  `gorm:"primary_key;auto_increment"`
	Username    string `gorm:"type:varchar(20); not null"`
	CouponName  string `gorm:"type:varchar(60); not null"`
	Amount      int64
	Left        int64
	Stock       float64
	Description string `gorm:"type:varchar(60)"`
}

type CouponRequest struct {
	Name        string
	Amount      string
	Description string
	Stock       string
}

type CouponResponse struct {
	Name        string  `json:"name"`
	Stock       float64 `json:"stock"`
	Description string  `json:"description"`
}

type SellerCouponRes struct {
	CouponResponse
	Amount int64 `json:"amount"`
	Left   int64 `json:"left"`
}

type CustomerCouponRes struct {
	CouponResponse
}

func ParseSellerCoupons(coupons []Coupon) []SellerCouponRes {
	var sellerCoupons []SellerCouponRes
	for _, coupon := range coupons {
		sellerCoupons = append(sellerCoupons,
			SellerCouponRes{CouponResponse{coupon.CouponName, coupon.Stock, coupon.Description},
				coupon.Amount, coupon.Left})
	}
	return sellerCoupons
}

func ParseCustomerCoupons(coupons []Coupon) []CustomerCouponRes {
	var sellerCoupons []CustomerCouponRes
	for _, coupon := range coupons {
		sellerCoupons = append(sellerCoupons,
			CustomerCouponRes{CouponResponse{coupon.CouponName, coupon.Stock, coupon.Description}})
	}
	return sellerCoupons
}
