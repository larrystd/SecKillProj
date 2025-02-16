package test

import (
	"SecKill/api"
	"SecKill/dao"
	"SecKill/engine"
	"SecKill/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
)

const registerUserPath = "/api/users"

func startServer(t *testing.T) (*httptest.Server, *httpexpect.Expect) {
	server := httptest.NewServer(engine.SeckillEngine())

	return server, httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),

		// use http.Client with a cookie jar and timeout
		Client: &http.Client{
			Jar:     httpexpect.NewJar(),
			Timeout: time.Second * 30,
		},
		// use verbose logging
		Printers: []httpexpect.Printer{
			httpexpect.NewCurlPrinter(t),
			httpexpect.NewDebugPrinter(t, true),
		},
	})
}

func testInvalidRegistrationAttempts(e *httpexpect.Expect) {
	validUser := "testuser"
	validPassword := "testpassword"

	// Test cases for invalid registration attempts
	testCases := []struct {
		name        string
		username    string
		password    string
		kind        string
		expectedMsg string
	}{
		{
			name:        "too short username",
			username:    "p",
			password:    validPassword,
			kind:        model.NormalSeller,
			expectedMsg: "User name too short.",
		},
		{
			name:        "too short password",
			username:    validUser,
			password:    "p",
			kind:        model.NormalSeller,
			expectedMsg: "Password too short.",
		},
		{
			name:        "empty user kind",
			username:    validUser,
			password:    validPassword,
			kind:        "",
			expectedMsg: "Empty field of kind.",
		},
		{
			name:        "invalid user kind",
			username:    validUser,
			password:    validPassword,
			kind:        "ImpossibleKind",
			expectedMsg: "Unexpected value of kind, ImpossibleKind",
		},
	}

	for _, tc := range testCases {
		e.POST(registerUserPath).
			WithJSON(RegisterForm{
				Username: tc.username,
				Password: tc.password,
				Kind:     tc.kind,
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ValueEqual(api.ErrMsgKey, tc.expectedMsg)
	}
}

func testDuplicateRegistration(e *httpexpect.Expect) {
	// Test successful registration
	e.POST(registerUserPath).
		WithJSON(RegisterForm{
			Username: "kiana",
			Password: "shen6508",
			Kind:     model.NormalSeller,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ValueEqual(api.ErrMsgKey, "")

	// Test duplicate registration attempt
	e.POST(registerUserPath).
		WithJSON(RegisterForm{
			Username: "kiana",
			Password: "password2",
			Kind:     model.NormalSeller,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().
		ValueEqual(api.ErrMsgKey, "Insert user failed. Maybe user name duplicates.")
}

func TestRegistrationScenarios(t *testing.T) {
	_, e := startServer(t)
	defer dao.Close()

	// Execute test scenarios
	testInvalidRegistrationAttempts(e)
	testDuplicateRegistration(e)
}
