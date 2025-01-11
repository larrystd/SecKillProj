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

type ReqCoupon struct {
	Name        string
	Amount      string
	Description string
	Stock       string
}

type ResCoupon struct {
	Name        string  `json:"name"`
	Stock       float64 `json:"stock"`
	Description string  `json:"description"`
}

type SellerResCoupon struct {
	ResCoupon
	Amount int64 `json:"amount"`
	Left   int64 `json:"left"`
}

type CustomerResCoupon struct {
	ResCoupon
}

func ParseSellerResCoupons(coupons []Coupon) []SellerResCoupon {
	var sellerCoupons []SellerResCoupon
	for _, coupon := range coupons {
		sellerCoupons = append(sellerCoupons,
			SellerResCoupon{ResCoupon{coupon.CouponName, coupon.Stock, coupon.Description},
				coupon.Amount, coupon.Left})
	}
	return sellerCoupons
}

func ParseCustomerResCoupons(coupons []Coupon) []CustomerResCoupon {
	var sellerCoupons []CustomerResCoupon
	for _, coupon := range coupons {
		sellerCoupons = append(sellerCoupons,
			CustomerResCoupon{ResCoupon{coupon.CouponName, coupon.Stock, coupon.Description}})
	}
	return sellerCoupons
}
