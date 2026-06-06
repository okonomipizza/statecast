package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/okonomipizza/statecast/internal/store"
)

func TestHandlePutStateRejectsOversizedBody(t *testing.T) {
	// 上限を超えるリクエストボディは拒否されること
	s := NewServer(store.New())
	if _, err := s.store.Register("agent"); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	body := bytes.Repeat([]byte("a"), maxRequestBodySize+1)
	req := httptest.NewRequest(http.MethodPut, "/v1/agents/agent/state", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	s.handlePutState(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status code mismatch: got %d", rec.Code)
	}
}

func TestHandleRegisterRejectsOversizedBody(t *testing.T) {
	// register も上限を超えるボディを拒否すること
	s := NewServer(store.New())

	body := strings.Repeat("a", maxRequestBodySize+1)
	req := httptest.NewRequest(http.MethodPost, "/v1/agents", strings.NewReader(body))
	rec := httptest.NewRecorder()

	s.handleRegister(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status code mismatch: got %d", rec.Code)
	}
}

func TestReadLimitedBody(t *testing.T) {
	// readLimitedBody は limit 以内のみ受け付けること
	body, err := readLimitedBody(strings.NewReader("ok"), 10)
	if err != nil || string(body) != "ok" {
		t.Fatalf("unexpected result: body=%q err=%v", body, err)
	}

	_, err = readLimitedBody(strings.NewReader(strings.Repeat("x", 11)), 10)
	if err != errRequestBodyTooLarge {
		t.Fatalf("expected errRequestBodyTooLarge, got %v", err)
	}
}
