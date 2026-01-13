package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"qd-sc/internal/model"
)

func TestLLMClient_ChatCompletion_RetryRebuildsRequestBody(t *testing.T) {
	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if len(body) == 0 {
			t.Fatalf("expected non-empty request body")
		}

		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 第二次返回一个最小合法响应
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(model.ChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
			Choices: []model.Choice{{Index: 0, FinishReason: "stop"}},
		})
	}))
	defer srv.Close()

	c := &LLMClient{
		baseURL:    srv.URL,
		apiKey:     "test",
		httpClient: srv.Client(),
		maxRetries: 2,
	}

	_, err := c.ChatCompletion(&model.ChatCompletionRequest{
		Model: "test-model",
		Messages: []model.Message{
			{Role: "user", Content: "hi"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls due to retry, got %d", got)
	}
}
