//go:build integration
// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

type submitReq struct {
	Type         string      `json:"type"`
	Priority     int         `json:"priority"`
	ThreadDemand int         `json:"thread_demand"`
	Payload      interface{} `json:"payload"`
}

type jobResp struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Priority     int         `json:"priority"`
	ThreadDemand int         `json:"thread_demand"`
	Status       string      `json:"status"`
	Result       interface{} `json:"result"`
}

func apiBase() string {
	if v := os.Getenv("API_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

func submitJob(t *testing.T, req submitReq) jobResp {
	t.Helper()
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal submit: %v", err)
	}
	resp, err := http.Post(apiBase()+"/jobs", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("submit request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("submit returned status %d: %s", resp.StatusCode, string(body))
	}
	var jr jobResp
	if err := json.NewDecoder(resp.Body).Decode(&jr); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}
	return jr
}

func pollJob(t *testing.T, id string, timeout time.Duration) jobResp {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("%s/jobs/%s", apiBase(), id))
		if err != nil {
			// retry
			time.Sleep(250 * time.Millisecond)
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			// maybe not yet written to redis/in-memory
			time.Sleep(250 * time.Millisecond)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			t.Fatalf("unexpected status %d polling job: %s", resp.StatusCode, string(body))
		}
		var jr jobResp
		if err := json.NewDecoder(resp.Body).Decode(&jr); err != nil {
			resp.Body.Close()
			t.Fatalf("decode poll response: %v", err)
		}
		resp.Body.Close()
		if jr.Status == "Completed" || jr.Status == "Failed" {
			return jr
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for job %s", id)
	return jobResp{}
}

func TestAPI_Smoke_AllJobs(t *testing.T) {
	// AddNumbers
	aj := submitJob(t, submitReq{Type: "add_numbers", Priority: 1, ThreadDemand: 1, Payload: map[string]interface{}{"x": 2, "y": 3}})
	ajr := pollJob(t, aj.ID, 10*time.Second)
	if ajr.Status != "Completed" {
		t.Fatalf("add_numbers not completed: %v", ajr)
	}

	// ReverseString
	rj := submitJob(t, submitReq{Type: "reverse_string", Priority: 1, ThreadDemand: 1, Payload: map[string]interface{}{"text": "hello"}})
	rjr := pollJob(t, rj.ID, 10*time.Second)
	if rjr.Status != "Completed" {
		t.Fatalf("reverse_string not completed: %v", rjr)
	}
	// verify result contains reversed string
	if m, ok := rjr.Result.(map[string]interface{}); ok {
		if rev, ok := m["reversed"].(string); !ok || rev != "olleh" {
			t.Fatalf("unexpected reverse result: %v", rjr.Result)
		}
	} else {
		t.Fatalf("unexpected reverse result shape: %T", rjr.Result)
	}

	// ResizeImage
	ir := submitJob(t, submitReq{Type: "resize_image", Priority: 1, ThreadDemand: 1, Payload: map[string]interface{}{"url": "http://ex/1.png", "width": 10, "height": 20}})
	irr := pollJob(t, ir.ID, 10*time.Second)
	if irr.Status != "Completed" {
		t.Fatalf("resize_image not completed: %v", irr)
	}

	// LargeArraySum
	lj := submitJob(t, submitReq{Type: "large_array_sum", Priority: 1, ThreadDemand: 2, Payload: map[string]interface{}{"array": []int{1, 2, 3, 4}}})
	ljr := pollJob(t, lj.ID, 10*time.Second)
	if ljr.Status != "Completed" {
		t.Fatalf("large_array_sum not completed: %v", ljr)
	}
	if m, ok := ljr.Result.(map[string]interface{}); ok {
		if sum, ok := m["sum"].(float64); !ok || int(sum) != 10 {
			t.Fatalf("unexpected large_array_sum result: %v", ljr.Result)
		}
	} else {
		t.Fatalf("unexpected large_array_sum result shape: %T", ljr.Result)
	}
}
