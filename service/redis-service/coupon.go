package redisService

import (
	"SecKill/dao"
	"SecKill/model"
	dbService "SecKill/service/db-service"
	"fmt"
	"log"
	"strconv"
)

func generateUserCouponKey(userName string) string {
	return fmt.Sprintf("%s-has", userName)
}

func generateCouponInfoKey(couponName string) string {
	return fmt.Sprintf("%s-info", couponName)
}

func AddUserCouponToRedis(coupon model.Coupon) (int64, error) {
	key := generateUserCouponKey(coupon.Username) //得到的key其实就是 coupon.Username-has
	val, err := dao.SetAdd(key, coupon.CouponName)
	return val, err
}

func UpdateCouponInfoToRedis(coupon model.Coupon) (string, error) {
	key := generateCouponInfoKey(coupon.CouponName)
	fields := map[string]interface{}{
		"id":          coupon.Id,
		"username":    coupon.Username,
		"couponName":  coupon.CouponName,
		"amount":      coupon.Amount,
		"left":        coupon.Left,
		"stock":       coupon.Stock,
		"description": coupon.Description,
	}
	val, err := dao.SetMapForever(key, fields)
	return val, err
}

func UpdateCouponCache(coupon model.Coupon) error {
	if _, err := AddUserCouponToRedis(coupon); err != nil {
		return err
	}

	if user, err := dbService.GetUser(coupon.Username); err != nil {
		log.Println("Database service error: ", err)
		return err
	} else {
		if user.IsSeller() {
			_, err = UpdateCouponInfoToRedis(coupon)
		}
		return err
	}
}

func GetCouponFromRedis(couponName string) model.Coupon {
	key := generateCouponInfoKey(couponName)
	values, err := dao.GetMap(key, "id", "username", "couponName", "amount", "left", "stock", "description")
	if err != nil {
		log.Println("Error on getting coupon. " + err.Error())
	}

	id, err := strconv.ParseInt(values[0].(string), 10, 64)
	if err != nil {
		log.Println("Wrong type of id. " + err.Error())
	}
	amount, err := strconv.ParseInt(values[3].(string), 10, 64)
	if err != nil {
		log.Println("Wrong type of amount. " + err.Error())
	}
	left, err := strconv.ParseInt(values[4].(string), 10, 64)
	if err != nil {
		log.Println("Wrong type of left. " + err.Error())
	}
	stock, err := strconv.ParseFloat(values[5].(string), 64)
	if err != nil {
		log.Println("Wrong type of stock. " + err.Error())
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
	hasCouponsKey := generateUserCouponKey(userName)
	couponNames, err := dao.GetSetMembers(hasCouponsKey)
	if err != nil {
		println("Error when getting coupon members. " + err.Error())
		return nil, err
	}

	for _, couponName := range couponNames {
		coupon := GetCouponFromRedis(couponName)
		coupons = append(coupons, coupon)
	}
	return coupons, nil
}
