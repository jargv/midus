package plumbus

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReturnStruct(t *testing.T) {
	type Result struct {
		Message string
	}

	handler := HandlerFunc(func() Result {
		return Result{
			Message: "Victory!",
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("couldn't get: %#v\n", err)
	}

	var result Result
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&result)
	if err != nil {
		t.Errorf("couldn't decode: %#v\n", err)
	}

	if result.Message != "Victory!" {
		t.Errorf(`body != "Victory!", body == %q`, result.Message)
	}
}

func TestReturnError(t *testing.T) {
	handler := HandlerFunc(func() (string, error) {
		return "", Errorf(http.StatusBadRequest, "result")
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("couldn't get: %#v\n", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("statusCode != 'http.StatusBadRequest', statusCode == %d", resp.StatusCode)
	}
}

func TestRequestBody(t *testing.T) {
	type Body struct {
		Message string
	}

	var message string

	handler := HandlerFunc(func(body *Body) {
		message = body.Message
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	bytes := bytes.Buffer{}
	json.NewEncoder(&bytes).Encode(&Body{
		Message: "full circle!",
	})

	_, err := http.Post(server.URL, "", &bytes)
	if err != nil {
		t.Errorf("couldn't make request: %#v\n", err)
	}

	if message != "full circle!" {
		t.Errorf(`message != "full circle", message == %q`, message)
	}
}

type Param string

func (p *Param) FromRequest(req *http.Request) error {
	*p = Param(req.URL.Query().Get("param"))
	return nil
}

func TestRequestParam(t *testing.T) {
	var param1 string
	var param2 string

	handler := HandlerFunc(func(p1 Param, p2 *Param) {
		param1 = string(p1)
		param2 = string(*p2)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	http.Get(server.URL + `?param=awesome`)

	if param1 != "awesome" {
		t.Errorf(`param1 != "awesome", param1 == %q`, param1)
	}

	if param2 != "awesome" {
		t.Errorf(`param2 != "awesome", param2 == %q`, param2)
	}
}

func TestRequestMethod(t *testing.T) {
	handler := HandlerFunc(&ByMethod{
		PUT: HandlerFunc(func() string {
			return "nachos"
		}),
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Errorf(": %#v\n", err)
	}

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf(
			`StatusCode != http.StatusMethodNotAllowed, StatusCode = %s`,
			resp.Status,
		)
	}

	req, err := http.NewRequest("PUT", server.URL, nil)
	if err != nil {
		t.Errorf("couldn't make request: %#v\n", err)
	}

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("couldn't make request: %#v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf(
			`StatusCode != http.StatusOK, StatusCode = %s`,
			resp.Status,
		)
	}
}

type userId string

func (ui *userId) FromRequest(req *http.Request) error {
	*ui = userId(req.URL.Query().Get("userId"))
	return nil
}

func TestPathParams(t *testing.T) {
	var result string
	mux := NewServeMux()
	mux.Handle("/user/:userId/name", func(id userId) {
		result = string(id)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	_, err := http.Get(server.URL + "/user/10/name")
	if err != nil {
		t.Fatalf("couldn't make request: %v\n", err)
	}

	if result != "10" {
		t.Fatalf(`result != "10", result == "%v"`, result)
	}
}

func TestRequiredRequestParam(t *testing.T) {
	type foodQueryParam string
	type amountQueryParam int
	var result string
	var amount int
	server := httptest.NewServer(HandlerFunc(func(food foodQueryParam, a amountQueryParam) {
		amount = int(a)
		result = string(food)
	}))

	//test that it's required (we should get a StatusBadRequest)
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("making request: %v\n", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf(`resp.StatusCode != htp.StatusBadRequest, resp.StatusCode == "%v"`, resp.StatusCode)
	}

	//test that it's required (we should get a StatusBadRequest)
	resp, err = http.Get(server.URL + "?food=nachos")
	if err != nil {
		t.Fatalf("making request: %v\n", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf(`resp.StatusCode != htp.StatusBadRequest, resp.StatusCode == "%v"`, resp.StatusCode)
	}

	//test that it's converted
	_, err = http.Get(server.URL + "?food=nachos&amount=10")
	if err != nil {
		t.Fatalf("makeing request: %v\n", err)
	}

	if result != "nachos" {
		t.Fatalf(`result != "nachos", result == "%v"`, result)
	}

	if amount != 10 {
		t.Fatalf(`amount != 10, amount == "%v"`, amount)
	}
}

func TestOptionalRequestParam(t *testing.T) {
	type foodQueryParam string
	type amountQueryParam int
	result := "not set"
	amount := 0
	server := httptest.NewServer(HandlerFunc(func(a *amountQueryParam, food *foodQueryParam) {
		if food != nil {
			result = string(*food)
		}
		if a != nil {
			amount = int(*a)
		}
	}))

	//test that it's not required
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("makeing request: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`resp.StatusCode != http.StatusOK, resp.StatusCode == "%v"`, resp.StatusCode)
	}

	if result != "not set" {
		t.Fatalf(`result != "not set", result == "%v"`, result)
	}

	//test that it's not required (on the second, int param)
	resp, err = http.Get(server.URL + "?food=nachos")
	if err != nil {
		t.Fatalf("makeing request: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(`resp.StatusCode != http.StatusOK, resp.StatusCode == "%v"`, resp.StatusCode)
	}

	if result != "nachos" {
		t.Fatalf(`result != "not set", result == "%v"`, result)
	}

	if amount != 0 {
		t.Fatalf(`amount != 0, amount == "%v"`, amount)
	}

	//test that it's passed to the handler
	_, err = http.Get(server.URL + "?food=nachos&amount=10")
	if err != nil {
		t.Fatalf("makeing request: %v\n", err)
	}

	if result != "nachos" {
		t.Fatalf(`result != "nachos", result == "%v"`, result)
	}

	if amount != 10 {
		t.Fatalf(`amount != 10, amount == "%v"`, amount)
	}
}

// type UserId struct {
// }

// func (ui *UserId) FromRequest(req *http.Request) error {
// 	return nil
// }

// type User struct {
// 	Name string `json:"name"`
// 	Age  int    `json:"age"`
// }

// type UserRepo struct {
// }

// func (ur *UserRepo) FindById(id UserId) (*User, error) {
// 	return nil, nil
// }

// func (ur *UserRepo) Edit(id UserId, user *User) error {
// 	return nil
// }

// func TestDocumentation(t *testing.T) {
// 	mux := NewServeMux()
// 	type user struct {
// 		Name string
// 		Age  int
// 	}

// 	type result struct {
// 		Role   string
// 		Id     int
// 		User   *user
// 		Thing1 *int
// 		Thing2 []int
// 		Thing3 []*int
// 		Thing4 []**int
// 		Thing5 map[string]*user
// 	}

// 	users := UserRepo{}

// 	mux.Handle("/users/:userId/details", func(u user) *result {
// 		return nil
// 	})

// 	mux.Handle("/users/:userId", ByMethod{
// 		GET: users.FindById,
// 		PUT: users.Edit,
// 	})

// 	mux.Handle("/standerd/handler", func(http.ResponseWriter, *http.Request) {})

// 	mux.Handle("/any/body", func(interface{}) {})

// 	docs := mux.Documentation()

// 	bytes, _ := json.MarshalIndent(docs, "", "  ")
// 	log.Printf("string(bytes):\n%s", string(bytes))
// }
