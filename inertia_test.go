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

func TestPartialReload(t *testing.T) {
	tests := []struct {
		name                   string
		component              string
		partialComponentHeader string
		partialDataHeader      string
		props                  P
		expectedProps          P
	}{
		{
			"Send only requested data",
			"events",
			"events",
			"partial",
			P{
				"foo":     "bar",
				"partial": "data",
			},
			P{
				"partial": "data",
			},
		},
		{
			"Rendered component differs from partial header",
			"notEvents",
			"events",
			"partial",
			P{
				"foo":     "bar",
				"partial": "data",
			},
			P{
				"foo":     "bar",
				"partial": "data",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			req, _ := http.NewRequest(http.MethodGet, "test", nil)
			req.Header.Add("X-Inertia", "true")
			req.Header.Add("X-Inertia-Partial-Data", tc.partialDataHeader)
			req.Header.Add("X-Inertia-Partial-Component", tc.partialComponentHeader)
			rw := httptest.NewRecorder()

			intertiaInstance := Inertia{Version: "test"}
			intertiaInstance.render(rw, req, tc.component, tc.props)

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

			if len(sentProps) != len(tc.expectedProps) {
				t.Errorf("Expected props %s, got %s", tc.expectedProps, sentProps)
			}

			for k := range tc.expectedProps {
				if tc.expectedProps[k] != sentProps[k] {
					t.Errorf("Expected props %s, got %s", tc.expectedProps, sentProps)
				}
			}

		})
	}
}
