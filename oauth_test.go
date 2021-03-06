package strava

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOAuthCallbackHandler(t *testing.T) {
	originalHttpClient := httpClient
	defer func() {
		httpClient = originalHttpClient
	}()

	// access denied
	f := OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
		t.Error("access denied should be failure")
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		if err != OAuthAuthorizationDeniedErr {
			t.Errorf("returned incorret error, got %v", err)
		}
	})

	req, _ := http.NewRequest("GET", "?error=access_denied", nil)
	f(httptest.NewRecorder(), req)

	// http client failure
	f = OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
		t.Error("should handle request failure")
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		if err == nil {
			t.Error("error should not be nil")
		}
	})

	httpClient = &http.Client{Transport: &storeRequestTransport{}}
	req, _ = http.NewRequest("GET", "", nil)

	f(httptest.NewRecorder(), req)

	// strava error
	f = OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
		t.Error("should return error when strava returned error")
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		if err != OAuthServerErr {
			t.Errorf("returned incorrect error, got %v", err)
		}
	})

	httpClient = NewStubResponseClient("{}", http.StatusInternalServerError).httpClient
	req, _ = http.NewRequest("GET", "", nil)

	f(httptest.NewRecorder(), req)

	// strava error
	f = OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
		t.Error("should return error when strava returned error")
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		if err != OAuthServerErr {
			t.Errorf("returned incorrect error, got %v", err)
		}
	})

	httpClient = NewStubResponseClient(`{"message":"bad","errors":[]}`, http.StatusBadRequest).httpClient
	req, _ = http.NewRequest("GET", "", nil)

	f(httptest.NewRecorder(), req)

	// strava invalid credentials error
	f = OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
		t.Error("should return error when strava returned error")
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		if err != OAuthInvalidCredentialsErr {
			t.Errorf("returned incorrect error, got %v", err)
		}
	})

	httpClient = NewStubResponseClient(`{"message":"bad","errors":[{"resource":"Application","field":"","code":""}]}`, http.StatusBadRequest).httpClient
	req, _ = http.NewRequest("GET", "", nil)

	f(httptest.NewRecorder(), req)

	// strava invalid code error
	f = OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
		t.Error("should return error when strava returned error")
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		if err != OAuthInvalidCodeErr {
			t.Errorf("returned incorrect error, got %v", err)
		}
	})

	httpClient = NewStubResponseClient(`{"message":"bad","errors":[{"resource":"RequestToken","field":"","code":""}]}`, http.StatusBadRequest).httpClient
	req, _ = http.NewRequest("GET", "", nil)

	f(httptest.NewRecorder(), req)

	// other strava error
	f = OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
		t.Error("should return error when strava returned error")
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		if _, ok := err.(*Error); !ok {
			t.Errorf("returned incorrect error, got %v", err)
		}
	})

	httpClient = NewStubResponseClient(`{"message":"bad","errors":[{"resource":"Othere","field":"","code":""}]}`, http.StatusBadRequest).httpClient
	req, _ = http.NewRequest("GET", "", nil)

	f(httptest.NewRecorder(), req)

	// bad json
	f = OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
		t.Error("should return error when strava returned error")
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		if err == nil {
			t.Error("error should not be nil")
		}
	})

	httpClient = NewStubResponseClient(`bad json`, http.StatusOK).httpClient
	req, _ = http.NewRequest("GET", "", nil)

	f(httptest.NewRecorder(), req)

	// success!
	f = OAuthCallbackHandler(func(auth *AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	}, func(err error, w http.ResponseWriter, r *http.Request) {
		t.Error("should be success")
	})

	httpClient = NewStubResponseClient(`{}`, http.StatusOK).httpClient
	req, _ = http.NewRequest("GET", "", nil)

	f(httptest.NewRecorder(), req)
}

func TestOAuthCallbackPath(t *testing.T) {
	_, err := OAuthCallbackPath()
	if err == nil {
		t.Error("should return error since callback url is not set")
	}

	OAuthCallbackURL = "http://www.strava.c%om/"
	_, err = OAuthCallbackPath()
	if err == nil {
		t.Error("should return error since not a callback url")
	}

	OAuthCallbackURL = "http://abc.com/strava/oauth"
	s, _ := OAuthCallbackPath()
	if s != "/strava/oauth" {
		t.Error("incorrect path")
	}
}

func TestOAuthAuthorizationURL(t *testing.T) {
	var url string
	OAuthCallbackURL = "http://abc.com/strava/oauth"

	url = OAuthAuthorizationURL("state", Permissions.Public, false)
	if url != basePath+"/oauth/authorize?client_id=0&response_type=code&redirect_uri=http://abc.com/strava/oauth&scope=public&state=state" {
		t.Errorf("incorrect oauth url, got %v", url)
	}

	url = OAuthAuthorizationURL("state", Permissions.Public, true)
	if url != basePath+"/oauth/authorize?client_id=0&response_type=code&redirect_uri=http://abc.com/strava/oauth&scope=public&state=state&approval_prompt=force" {
		t.Errorf("incorrect oauth url, got %v", url)
	}

	url = OAuthAuthorizationURL("state", Permissions.ViewPrivate, false)
	if url != basePath+"/oauth/authorize?client_id=0&response_type=code&redirect_uri=http://abc.com/strava/oauth&scope=view_private&state=state" {
		t.Errorf("incorrect oauth url, got %v", url)
	}

	url = OAuthAuthorizationURL("", Permissions.Public, false)
	if url != basePath+"/oauth/authorize?client_id=0&response_type=code&redirect_uri=http://abc.com/strava/oauth&scope=public" {
		t.Errorf("incorrect oauth url, got %v", url)
	}
}

func TestOAuthErrorError(t *testing.T) {
	err := OAuthAuthorizationDeniedErr
	if err.Error() != err.message {
		t.Error("should simply print message")
	}
}

func TestErrorString(t *testing.T) {
	err := Error{
		Message: "bad bad bad",
		Errors:  []*ErrorDetailed{&ErrorDetailed{"auth", "code", "missing"}},
	}
	if err.Error() != `{"message":"bad bad bad","errors":[{"resource":"auth","field":"code","code":"missing"}]}` {
		t.Errorf("should simply print message, got %v", err.Error())
	}
}
