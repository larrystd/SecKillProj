package redisService

import (
	"SecKill/dao"
	dbService "SecKill/service/db-service"
	"log"
)

const secKillScript = `
    -- Check if User has coupon
    -- KEYS[1]: hasCouponKey "{username}-has"
    -- KEYS[2]: couponName   "{couponName}"
    -- KEYS[3]: couponKey    "{couponName}-info"

    -- Check if coupon left
	local couponLeft = redis.call("hget", KEYS[3], "left");
	if (couponLeft == false)
	then
		return -2;  -- No such coupon
	end
	if (tonumber(couponLeft) == 0)
    then
		return -3;  --  No Coupon Left.
	end

    -- Check if the user has got the coupon --
	local userHasCoupon = redis.call("SISMEMBER", KEYS[1], KEYS[2]);
	if (userHasCoupon == 1)
	then
		return -1;
	end

    -- Let User get a coupon --
	redis.call("hset", KEYS[3], "left", couponLeft - 1);
	redis.call("SADD", KEYS[1], KEYS[2]);
	return 1;
`

var secKillSHA string // SHA expression of secKillScript

func preHeatRedis() {
	coupons, err := dbService.GetCouponsInfoFromDB()
	if err != nil {
		panic("Error when getting all coupons." + err.Error())
	}

	for _, coupon := range coupons {
		err := UpdateCouponCache(coupon)
		if err != nil {
			panic("Error while setting redis keys of coupons. " + err.Error())
		}
	}
	log.Println("---Set redis keys of coupons success.---")
}

func init() {
	secKillSHA = dao.PrepareScript(secKillScript)
	preHeatRedis()
}
