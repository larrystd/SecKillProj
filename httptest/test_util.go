package httptest

import (
	"SecKill/api"
	"SecKill/model"
	"fmt"
	"net/http"

	"github.com/gavv/httpexpect"
)

const addCouponPath = "/api/users/{username}/coupons"
const getCouponPath = "/api/users/{username}/coupons"
const fetchCouponPath = "/api/users/{username}/coupons/{name}"

const pageQueryKey = "page"

const demoSellerName = "kiana"
const demoCustomerName = "jinzili"
const demoArCustomerName = "karsa" // name of another customer
const demoPassword = "shen6508"

const demoCouponName = "my_coupon"
const demoAmount = 10
const demoStock = 50

type RegisterForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
	Kind     string `form:"kind"`
}

type LoginForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

type AddCouponForm struct {
	Name        string `form:"name"`
	Amount      string `form:"amount"`
	Description string `form:"description"`
	Stock       string `form:"stock"`
}

func registerDemoUsers(e *httpexpect.Expect) {
	e.POST("/api/users").
		WithJSON(RegisterForm{demoSellerName, demoPassword, model.NormalSeller}).
		Expect().
		Status(http.StatusOK).JSON().Object().
		ValueEqual(api.ErrMsgKey, "")

	e.POST("/api/users").
		WithJSON(RegisterForm{demoCustomerName, demoPassword, model.NormalCustomer}).
		Expect().
		Status(http.StatusOK).JSON().Object().
		ValueEqual(api.ErrMsgKey, "")

	e.POST("/api/users").
		WithJSON(RegisterForm{demoArCustomerName, demoPassword, model.NormalCustomer}).
		Expect().
		Status(http.StatusOK).JSON().Object().
		ValueEqual(api.ErrMsgKey, "")

}

func doLogin(e *httpexpect.Expect, username, password, kind string) string {
	resp := e.POST("/api/auth").
		WithJSON(LoginForm{username, password}).
		Expect().
		Status(http.StatusOK)

	authToken := resp.Header("Authorization").Raw()

	resp.JSON().Object().
		ValueEqual(api.ErrMsgKey, "").
		ValueEqual("kind", kind)
	return authToken
}

func logout(e *httpexpect.Expect) {
	e.POST("/api/auth/logout").
		Expect().
		Status(http.StatusOK).JSON().Object().
		ValueEqual(api.ErrMsgKey, "log out.")
}

var customerSchema = fmt.Sprintf(`{
	"type": "object",
	"properties": {
		"%s": {
				"type": "string"
			},
        "%s": {
				"type": "array",
				"items": {
					"type":        "object",
					"name":        "string",
					"amount":      "integer",
					"left":        "integer",
					"stock":       "integer",
					"description": "string"
				}
			}
	}
}`, api.ErrMsgKey, api.DataKey)

var sellerSchema = fmt.Sprintf(`{
	"type": "object",
	"properties": {
		"%s": {
				"type": "string"
			},
        "%s": {
				"type": "array",
				"items": {
					"type":        "object",
					"name":        "string",
					"stock":       "float64",
					"description": "string"
				}
			}
	}
}`, api.ErrMsgKey, api.DataKey)
