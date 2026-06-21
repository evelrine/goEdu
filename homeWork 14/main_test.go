package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHelloHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/hello", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    helloHandler(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }

    contentType := rr.Header().Get("Content-Type")
    expectedContentType := "application/json"
    if contentType != expectedContentType {
        t.Errorf("handler returned wrong Content-Type: got %v want %v",
            contentType, expectedContentType)
    }

    var response map[string]string
    err = json.Unmarshal(rr.Body.Bytes(), &response)
    if err != nil {
        t.Errorf("response body is not valid JSON: %v", err)
    }

    message, exists := response["message"]
    if !exists {
        t.Error("response JSON does not contain 'message' field")
    }

    expectedMessage := "hello, world!"
    if message != expectedMessage {
        t.Errorf("message value is wrong: got %v want %v",
            message, expectedMessage)
    }
}
