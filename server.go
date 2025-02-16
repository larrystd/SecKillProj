package main

import (
	"SecKill/dao"
	"SecKill/engine"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

const port = 20080

func main() {
	router := engine.SeckillEngine()
	defer dao.Close()

	go func() {
		log.Println("pprof start...")
		log.Println(http.ListenAndServe(":9876", nil))
	}()

	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Println("Error when running server. " + err.Error())
	}
}
