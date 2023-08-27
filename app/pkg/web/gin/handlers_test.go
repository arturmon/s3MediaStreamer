package gin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	_ "net/http"
	"net/http/httptest"
	_ "net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	router := gin.Default()
	router.GET("/ping", Ping)

	router.ServeHTTP(rec, req)

	expected := `{"message":"pong"}`
	received := rec.Body.String()

	var expectedJSON, receivedJSON interface{}

	// Unmarshal the expected and received JSON to interface{}
	err := json.Unmarshal([]byte(expected), &expectedJSON)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal([]byte(received), &receivedJSON)
	if err != nil {
		t.Fatal(err)
	}

	// Marshal the interface{} back to JSON with compact formatting
	expectedBytes, _ := json.Marshal(expectedJSON)
	receivedBytes, _ := json.Marshal(receivedJSON)

	if string(expectedBytes) != string(receivedBytes) {
		t.Errorf("expected %s must equal to received %s", expected, received)
	}
}
