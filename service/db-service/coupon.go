package dbService

import (
	"SecKill/dao"
	"SecKill/model"
	"fmt"
)

func GetCouponsInfoFromDB() ([]model.Coupon, error) {
	var coupons []model.Coupon
	result := dao.Db.Find(&coupons)
	return coupons, result.Error
}

func InsertCouponToCustomUser(userName string, coupon model.Coupon) error {
	return dao.Db.Exec(fmt.Sprintf("INSERT IGNORE INTO coupons "+
		"(`username`,`coupon_name`,`amount`,`left`,`stock`,`description`) "+
		"values('%s', '%s', %d, %d, %f, '%s')",
		userName, coupon.CouponName, 1, 1, coupon.Stock, coupon.Description)).Error
}

func DecreaseCouponLeftNum(sellerName string, couponName string) error {
	return dao.Db.Exec(fmt.Sprintf("UPDATE coupons c SET c.left=c.left-1 WHERE "+
		"c.username='%s' AND c.coupon_name='%s' AND c.left>0", sellerName, couponName)).Error
}
