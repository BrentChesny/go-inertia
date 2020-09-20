package inertia

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLazyLoading(t *testing.T) {
	props := P{
		"foo":  "bar",
		"lazy": func() interface{} { return "eval" },
	}
	req, _ := http.NewRequest(http.MethodGet, "test", nil)
	req.Header.Add("X-Inertia", "true")
	rw := httptest.NewRecorder()

	intertiaInstance := Inertia{Version: "test"}
	intertiaInstance.render(rw, req, "events", props)

	resp := rw.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(rw.Result().Body)
	unmarshalledBody := page{}
	err := json.Unmarshal(body, &unmarshalledBody)
	if err != nil {
		t.Error(err)
	}
	sentProps := unmarshalledBody.Props

	if !(len(sentProps) == 2 && sentProps["lazy"] == "eval") {
		t.Errorf("Expected props %v, got %s", props, sentProps)
	}
}

// TODO add test for partial reload
