package redisService

import (
	"SecKill/api/dbService"
	"SecKill/data"
	"SecKill/model"
	"fmt"
	"log"
	"strconv"
)

func getHasCouponsKeyByName(userName string) string {
	return fmt.Sprintf("%s-has", userName)
}

func getCouponKeyByCoupon(coupon model.Coupon) string {
	return getCouponKeyByName(coupon.CouponName)
}
func getCouponKeyByName(couponName string) string {
	return fmt.Sprintf("%s-info", couponName)
}

func CacheHasCoupon(coupon model.Coupon) (int64, error) {
	key := getHasCouponsKeyByName(coupon.Username) //得到的key其实就是 coupon.Username-has
	val, err := data.SetAdd(key, coupon.CouponName)
	return val, err
}

func CacheCoupon(coupon model.Coupon) (string, error) {
	key := getCouponKeyByCoupon(coupon)
	fields := map[string]interface{}{
		"id":          coupon.Id,
		"username":    coupon.Username,
		"couponName":  coupon.CouponName,
		"amount":      coupon.Amount,
		"left":        coupon.Left,
		"stock":       coupon.Stock,
		"description": coupon.Description,
	}
	val, err := data.SetMapForever(key, fields)
	return val, err
}

func CacheCouponAndHasCoupon(coupon model.Coupon) error {
	if _, err := CacheHasCoupon(coupon); err != nil {
		return err
	}

	if user, err := dbService.GetUser(coupon.Username); err != nil {
		log.Println("Database service error: ", err)
		return err
	} else {
		if user.IsSeller() {
			_, err = CacheCoupon(coupon)
		}
		return err
	}
}

func GetCoupon(couponName string) model.Coupon {
	key := getCouponKeyByName(couponName)
	values, err := data.GetMap(key, "id", "username", "couponName", "amount", "left", "stock", "description")
	if err != nil {
		println("Error on getting coupon. " + err.Error())
	}

	id, err := strconv.ParseInt(values[0].(string), 10, 64)
	if err != nil {
		println("Wrong type of id. " + err.Error())
	}
	amount, err := strconv.ParseInt(values[3].(string), 10, 64)
	if err != nil {
		println("Wrong type of amount. " + err.Error())
	}
	left, err := strconv.ParseInt(values[4].(string), 10, 64)
	if err != nil {
		println("Wrong type of left. " + err.Error())
	}
	stock, err := strconv.ParseFloat(values[5].(string), 64)
	if err != nil {
		println("Wrong type of stock. " + err.Error())
	}
	return model.Coupon{
		Id:          id,
		Username:    values[1].(string),
		CouponName:  values[2].(string),
		Amount:      amount,
		Left:        left,
		Stock:       stock,
		Description: values[6].(string),
	}

}

func GetCoupons(userName string) ([]model.Coupon, error) {
	var coupons []model.Coupon
	hasCouponsKey := getHasCouponsKeyByName(userName)
	couponNames, err := data.GetSetMembers(hasCouponsKey)
	if err != nil {
		println("Error when getting coupon members. " + err.Error())
		return nil, err
	}

	for _, couponName := range couponNames {
		coupon := GetCoupon(couponName)
		coupons = append(coupons, coupon)
	}
	return coupons, nil
}
