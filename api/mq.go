package api

import (
	"SecKill/model"
	dbService "SecKill/service/db-service"
	"log"
)

type secKillMessage struct {
	username string
	coupon   model.Coupon
}

const maxMessageNum = 20000

var secKillChannel = make(chan secKillMessage, maxMessageNum)

func seckillConsumer() {
	for {
		message := <-secKillChannel
		log.Println("Got one message: " + message.username)

		username := message.username
		sellerName := message.coupon.Username
		couponName := message.coupon.CouponName

		var err error
		err = dbService.InsertCouponToCustomUser(username, message.coupon) // update db
		if err != nil {
			log.Println("Error when inserting user's coupon. " + err.Error())
		}
		err = dbService.DecreaseCouponLeftNum(sellerName, couponName)
		if err != nil {
			log.Println("Error when decreasing coupon left. " + err.Error())
		}
	}
}

var isConsumerRun = false

func RunSecKillConsumer() {
	// Only Run one consumer.
	if !isConsumerRun {
		go seckillConsumer()
		isConsumerRun = true
	}
}
