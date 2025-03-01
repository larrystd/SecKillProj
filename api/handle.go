package api

import (
	"SecKill/dao"
	"SecKill/middleware/jwt"
	"SecKill/model"
	redisService "SecKill/service/redis-service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Visible for testing
const ErrMsgKey = "errMsg"
const DataKey = "data"

func FetchCoupon(ctx *gin.Context) {
	// check authorized
	claims := ctx.MustGet("claims").(*jwt.CustomClaims)
	if claims == nil {
		log.Printf("context claims is nil")
		ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Not Authorized."})
		return
	}
	if claims.Kind == "saler" {
		ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Sellers aren't allowed to get coupons."})
		return
	}

	paramSellerName := ctx.Param("username")
	paramCouponName := ctx.Param("name")
	// check coupon left
	if _, err := redisService.CacheAtomicSecKill(claims.Username, paramSellerName, paramCouponName); err != nil {
		if redisService.IsRedisEvalError(err) {
			log.Println("Server error" + err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{ErrMsgKey: err.Error()})
			return
		}
		log.Println("Fail to fetch coupon. " + err.Error())
		ctx.JSON(http.StatusNoContent, gin.H{})
		return
	}
	coupon := redisService.GetCouponFromRedis(paramCouponName)
	secKillChannel <- secKillMessage{claims.Username, coupon}
	ctx.JSON(http.StatusCreated, gin.H{ErrMsgKey: ""})
}

const (
	couponPageSize int64 = 20
)

func getValidCouponSlice(allCoupons []model.Coupon, page int64) []model.Coupon {
	if len(allCoupons) == 0 {
		return allCoupons
	}
	couponLen := int64(len(allCoupons))
	startIndex := page * couponPageSize
	endIndex := page*couponPageSize + couponPageSize
	if startIndex < 0 {
		startIndex = 0
	} else if startIndex > couponLen {
		startIndex = couponLen
	}
	if endIndex < 1 {
		if couponLen < couponPageSize {
			endIndex = couponLen
		} else {
			endIndex = couponPageSize
		}
	} else if endIndex > couponLen {
		endIndex = couponLen
	}
	return allCoupons[startIndex:endIndex]
}

func getDataStatusCode(len int) int {
	if len == 0 {
		return http.StatusNoContent
	}
	return http.StatusOK
}

func ListCoupons(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*jwt.CustomClaims)
	if claims == nil {
		log.Println("claims is nil.")
		ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Not Authorized."})
		return
	}
	queryUserName, queryPage := ctx.Param("username"), ctx.Query("page")

	var page int64
	var tmpPage int64
	if queryPage == "" {
		tmpPage = 1
	} else {
		var err error
		tmpPage, err = strconv.ParseInt(ctx.Query("page"), 10, 64)
		if err != nil {
			log.Println("Wrong format of page.")
			ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Wrong format of page."})
			return
		}
	}
	page = tmpPage - 1

	queryUser := model.User{Username: queryUserName}
	queryErr := dao.Db.Where(&queryUser).First(&queryUser).Error
	if queryErr != nil {
		if gorm.IsRecordNotFoundError(queryErr) {
			log.Println("Record not found.")
			ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Record not found."})
			return
		}
		log.Println("Query error.")
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Query error."})
		return
	}
	// custom query its coupons info
	if queryUserName == claims.Username {
		var allCoupons []model.Coupon
		var err error
		// get coupons from redis
		if allCoupons, err = redisService.GetCoupons(claims.Username); err != nil {
			log.Println("Server error.")
			ctx.JSON(http.StatusInternalServerError, gin.H{ErrMsgKey: "Server error."})
			return
		}
		// all coupons -> current page coupons
		coupons := getValidCouponSlice(allCoupons, page)

		if queryUser.IsSeller() {
			sellerCoupons := model.ParseSellerCoupons(coupons)
			statusCode := getDataStatusCode(len(sellerCoupons))
			ctx.JSON(statusCode, gin.H{ErrMsgKey: "", DataKey: sellerCoupons})
			return
		} else if queryUser.IsCustomer() {
			customerCoupons := model.ParseCustomerCoupons(coupons)
			statusCode := getDataStatusCode(len(customerCoupons))
			ctx.JSON(statusCode, gin.H{ErrMsgKey: "", DataKey: customerCoupons})
			return
		}
	}
	// seller query comsumer coupon info
	if queryUser.IsSeller() {
		var allCoupons []model.Coupon
		var err error
		if allCoupons, err = redisService.GetCoupons(queryUserName); err != nil {
			log.Println("Error when getting seller's coupons.")
			ctx.JSON(http.StatusInternalServerError, gin.H{ErrMsgKey: "Error when getting seller's coupons.", DataKey: allCoupons})
			return
		}
		coupons := getValidCouponSlice(allCoupons, page)

		sellerCoupons := model.ParseSellerCoupons(coupons)
		statusCode := getDataStatusCode(len(sellerCoupons))
		ctx.JSON(statusCode, gin.H{ErrMsgKey: "", DataKey: sellerCoupons})
		return
	}
	log.Printf("Username check failed, %v\n.", queryUserName)
	ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Cannot check other customer.", DataKey: []model.Coupon{}})
}

func AddCoupon(ctx *gin.Context) {
	claims := ctx.MustGet("claims").(*jwt.CustomClaims)
	if claims == nil {
		log.Println("Not Authorized.")
		ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Not Authorized."})
		return
	}

	if claims.Kind == "customer" {
		log.Println("Only sellers can create coupons.")
		ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Only sellers can create coupons."})
		return
	}

	paramUserName := ctx.Param("username")
	var couponRequest model.CouponRequest
	var err error
	if err := ctx.BindJSON(&couponRequest); err != nil {
		log.Println("Only receive JSON format.")
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Only receive JSON format."})
		return
	}
	couponName := couponRequest.Name
	formAmount := couponRequest.Amount
	description := couponRequest.Description
	formStock := couponRequest.Stock
	if claims.Username != paramUserName {
		log.Println("Cannot create coupons for other users.")
		ctx.JSON(http.StatusUnauthorized, gin.H{ErrMsgKey: "Cannot create coupons for other users."})
		return
	}

	var amount int64
	var stock float64
	if amount, err = strconv.ParseInt(formAmount, 10, 64); err != nil {
		log.Printf("Cannot convert formAmount to int64, formAmount %v", formAmount)
	}

	if stock, err = strconv.ParseFloat(formStock, 64); err != nil {
		log.Printf("Cannot convert formStock to float64, formStock %v", formStock)
	}

	coupon := model.Coupon{
		Username:    claims.Username,
		CouponName:  couponName,
		Amount:      amount,
		Left:        amount,
		Stock:       stock,
		Description: description,
	}

	err = dao.Db.Create(&coupon).Error
	if err != nil {
		log.Println("Create failed. Maybe (username,coupon name) duplicates")
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Create failed. Maybe (username,coupon name) duplicates"})
		return
	}

	if err = redisService.UppdateCouponCache(coupon); err != nil {
		log.Println("Create Cache failed. ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{ErrMsgKey: "Create Cache failed. " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{ErrMsgKey: ""})
}

func RegisterUser(ctx *gin.Context) {
	var postUser model.RegisterUser

	if err := ctx.BindJSON(&postUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Only receive JSON format."})
		return
	}
	// TODO: 使用参数校验库优化校验
	if len(postUser.Username) < model.MinUserNameLen {
		log.Println("User name too short.")
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "User name too short."})
		return
	}
	if len(postUser.Password) < model.MinPasswordLen {
		log.Println("Password too short.")
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Password too short."})
		return
	}
	if postUser.Kind == "" {
		log.Println("Empty field of kind.")
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Empty field of kind."})
		return
	}
	if !model.IsValidKind(postUser.Kind) {
		log.Println("Unexpected value of kind, ", postUser.Kind)
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Unexpected value of kind, " + postUser.Kind})
		return
	}

	user := model.User{Username: postUser.Username, Kind: postUser.Kind, Password: model.GetMD5(postUser.Password)}
	if dao.Db.Create(&user).Error == nil {
		ctx.JSON(http.StatusOK, gin.H{ErrMsgKey: ""})
	} else {
		log.Println("Insert user failed. Maybe user name duplicates.")
		ctx.JSON(http.StatusBadRequest, gin.H{ErrMsgKey: "Insert user failed. Maybe user name duplicates."})
	}
}
