package redisService

import (
	"SecKill/dao"
	"fmt"

	"github.com/prometheus/common/log"
)

type redisEvalError struct {
}

func (e redisEvalError) Error() string {
	return "Error when executing redisService eval."
}

type userHasCouponError struct {
	userName   string
	couponName string
}

func (e userHasCouponError) Error() string {
	return fmt.Sprintf("User %s has had coupon %s.", e.userName, e.couponName)
}

type noSuchCouponError struct {
	userName   string
	couponName string
}

func (e noSuchCouponError) Error() string {
	return fmt.Sprintf("Coupon %s created by %s doesn't exist.", e.couponName, e.userName)
}

type noCouponLeftError struct {
	userName   string
	couponName string
}

func (e noCouponLeftError) Error() string {
	return fmt.Sprintf("No Coupon %s created by %s left.", e.couponName, e.userName)
}

type CouponLeftResError struct {
	couponLeftRes interface{}
}

func (e CouponLeftResError) Error() string {
	switch e.couponLeftRes.(type) {
	case int:
		return fmt.Sprintf("Unexpected couponLeftRes Num: %v.", e.couponLeftRes)
	default:
		return fmt.Sprintf("couponLeftRes : %v with wrong type.", e.couponLeftRes)
	}
}

func IsRedisEvalError(err error) bool {
	switch err.(type) {
	case redisEvalError:
		return true
	default:
		return false
	}
}

func CacheAtomicSecKill(userName string, sellerName string, couponName string) (int64, error) {
	userHasCouponsKey := generateUserCouponKey(userName)
	couponKey := generateCouponInfoKey(couponName)
	res, err := dao.EvalSHA(secKillSHA, []string{userHasCouponsKey, couponName, couponKey})
	if err != nil {
		return -1, redisEvalError{}
	}

	couponLeftRes, ok := res.(int64)
	if !ok {
		return -1, redisEvalError{}
	}
	// error handle
	switch couponLeftRes {
	case -1:
		return -1, userHasCouponError{userName, couponName}
	case -2:
		return -1, noSuchCouponError{sellerName, couponName}
	case -3:
		return -1, noCouponLeftError{sellerName, couponName}
	case 1:
		return couponLeftRes, nil
	default:
		{
			log.Fatal("Unexpected return value.")
			return -1, CouponLeftResError{couponLeftRes}
		}
	}
}
