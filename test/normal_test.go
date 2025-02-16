package test

import (
	"SecKill/api"
	"SecKill/dao"
	"SecKill/model"
	"net/http"
	"strconv"
	"testing"

	"github.com/gavv/httpexpect"
)

const loginPath = "/api/users"

func testInvalidLoginAttempts(e *httpexpect.Expect) {
	wrongUserName := "wrongUserName"
	e.POST(loginPath).
		WithJSON(LoginForm{wrongUserName, "whatever_pw"}).
		Expect().
		Status(http.StatusBadRequest).JSON().Object().
		ValueNotEqual(api.ErrMsgKey, "No such queryUser.")

	wrongPassword := "sysucs515"
	e.POST(loginPath).
		WithJSON(LoginForm{demoSellerName, wrongPassword}).
		Expect().
		Status(http.StatusBadRequest).JSON().Object().
		ValueNotEqual(api.ErrMsgKey, "Password mismatched.")
}

func testUserAuthenticationFlow(e *httpexpect.Expect) {
	testInvalidLoginAttempts(e)
	doLogin(e, demoCustomerName, demoPassword, model.NormalCustomer)
}

func getCouponUnauthorized(e *httpexpect.Expect, username string, page int) {
	jsonObject := e.GET(listCouponUrl, username).
		WithQuery(pageQueryKey, page).
		Expect().
		Status(http.StatusUnauthorized).JSON().Object()
	jsonObject.Value(api.ErrMsgKey).Equal("Cannot check other customer.")
	jsonObject.Value(api.DataKey).Equal([]model.Coupon{})
}

func testCouponManagement(e *httpexpect.Expect) {
	// Test coupon lifecycle: creation, retrieval, and verification
	customerAuthToken := doLogin(e, demoCustomerName, demoPassword, model.NormalCustomer)

	verifyEmptyCouponList := func(username string, page, status int) {
		e.GET(listCouponUrl, username).
			WithHeader("Authorization", customerAuthToken).
			WithQuery(pageQueryKey, page).
			Expect().
			Status(status).Body().Empty()
	}

	// Verify empty coupon lists for various scenarios
	verifyEmptyCouponList(demoCustomerName, -1, http.StatusNoContent)
	verifyEmptyCouponList(demoCustomerName, 0, http.StatusNoContent)
	verifyEmptyCouponList(demoSellerName, -1, http.StatusNoContent)
	verifyEmptyCouponList(demoSellerName, 0, http.StatusNoContent)

	// Seller adds new coupon
	sellerAuthToken := doLogin(e, demoSellerName, demoPassword, model.NormalSeller)
	e.POST(addCouponUrl, demoSellerName).
		WithHeader("Authorization", sellerAuthToken).
		WithJSON(AddCouponForm{
			demoCouponName,
			strconv.Itoa(demoAmount),
			strconv.Itoa(demoStock),
			"kiana: this is my good coupon",
		}).
		Expect().
		Status(http.StatusCreated).JSON().Object().
		ValueEqual(api.ErrMsgKey, "")

	// Verify coupon list is no longer empty
	verifyNonEmptyCouponList := func(username string, page, status int) {
		e.GET(listCouponUrl, username).
			WithHeader("Authorization", customerAuthToken).
			WithQuery(pageQueryKey, page).
			Expect().
			Status(status).JSON().Object().
			ValueEqual(api.ErrMsgKey, "").
			Value("data").Array().Length().Gt(0)
	}
	verifyNonEmptyCouponList(demoSellerName, 0, http.StatusOK)

	// Verify coupon schema
	verifyCouponSchema := func(username string, page int) {
		e.GET(listCouponUrl, username).
			WithHeader("Authorization", customerAuthToken).
			WithQuery(pageQueryKey, page).
			Expect().JSON().Schema(sellerSchema)
	}
	verifyCouponSchema(demoSellerName, 0)
}

func testCouponAcquisition(e *httpexpect.Expect, initialCouponAmount int) {
	// Initialize user sessions
	customerAuthToken := doLogin(e, demoCustomerName, demoPassword, model.NormalCustomer)
	sellerAuthToken := doLogin(e, demoSellerName, demoPassword, model.NormalSeller)

	// Helper function to attempt coupon acquisition
	attemptCouponAcquisition := func(username string, couponName string, expectedStatus int) {
		e.PATCH(fetchCouponUrl, username, couponName).
			WithHeader("Authorization", customerAuthToken).
			Expect().
			Status(expectedStatus)
	}

	// Helper function to verify remaining coupon stock
	verifyRemainingCoupons := func(username string, page int, index int, expectedRemaining int) {
		e.GET(listCouponUrl, username).
			WithHeader("Authorization", sellerAuthToken).
			WithQuery(pageQueryKey, page).
			Expect().
			JSON().Object().
			Value(api.DataKey).Array().
			Element(index).Object().
			Value("left").Equal(expectedRemaining)
	}

	// Helper function to verify customer's coupon list
	verifyCustomerCoupons := func(page int, expectedStatus int) {
		e.GET(listCouponUrl, demoCustomerName).
			WithHeader("Authorization", customerAuthToken).
			WithQuery(pageQueryKey, page).
			Expect().
			Status(expectedStatus).
			JSON().Object().
			ValueEqual(api.ErrMsgKey, "").
			Value("data").Array().Length().Gt(0)
	}

	// Test initial coupon acquisition
	attemptCouponAcquisition(demoCustomerName, demoCouponName, http.StatusCreated)

	// Verify coupon stock decreased by 1
	verifyRemainingCoupons(demoSellerName, 0, 0, initialCouponAmount-1)

	// Verify customer can see the acquired coupon
	verifyCustomerCoupons(0, http.StatusOK)

	// Test duplicate coupon acquisition attempt
	attemptCouponAcquisition(demoCustomerName, demoCouponName, http.StatusNoContent)
}

// normal test
func TestNormal(t *testing.T) {
	_, e := startServer(t)
	defer dao.Close()

	// Setup test environment, register seller and customer
	registerDemoUsers(e)

	// Execute test scenarios
	testUserAuthenticationFlow(e)
	testCouponManagement(e)
	testCouponAcquisition(e, demoAmount)
}
