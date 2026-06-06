package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/okonomipizza/statecast/internal/runtime"
	"github.com/okonomipizza/statecast/internal/store"
)

const (
	// maxRequestBodySize は HTTP リクエストボディの上限（1 MiB）。
	maxRequestBodySize = 1 << 20
)

var errRequestBodyTooLarge = errors.New("request body too large")

// Server は statecast daemon の HTTP API を提供する。
type Server struct {
	store      *store.Store
	socketPath string
	pidPath    string
	listener   net.Listener
}

// NewServer は新しい daemon Server を返す。
func NewServer(st *store.Store) *Server {
	return &Server{store: st}
}

// Serve は Unix ソケット上で HTTP サーバを起動する。
func (s *Server) Serve() error {
	if _, err := runtime.EnsureDir(); err != nil {
		return err
	}

	socketPath, err := runtime.SocketPath()
	if err != nil {
		return err
	}
	pidPath, err := runtime.PIDPath()
	if err != nil {
		return err
	}

	s.socketPath = socketPath
	s.pidPath = pidPath

	if err := removeStaleSocket(socketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen unix socket: %w", err)
	}
	s.listener = listener

	// PID ファイルは 0o600（所有者のみ読み書き）とする。
	// 0644 だと同一ホスト上の他ユーザーが内容を書き換えられ、
	// stop 時に送る SIGTERM の宛先 PID を偽装される恐れがある。
	// ランタイムディレクトリ自体は 0700 だが、ファイル単位でも制限しておく。
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		_ = listener.Close()
		return fmt.Errorf("write pid file: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /v1/agents", s.handleRegister)
	mux.HandleFunc("GET /v1/agents", s.handleList)
	mux.HandleFunc("GET /v1/agents/{name}/state", s.handleGetState)
	mux.HandleFunc("PUT /v1/agents/{name}/state", s.handlePutState)

	server := &http.Server{Handler: mux}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	select {
	case <-sigCh:
		_ = server.Close()
		_ = s.Close()
		return nil
	case err := <-serveErr:
		_ = s.Close()
		if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
			return nil
		}
		return err
	}
}

// Close はリスナーとランタイムファイルを片付ける。
func (s *Server) Close() error {
	var closeErr error
	if s.listener != nil {
		closeErr = s.listener.Close()
	}
	if s.socketPath != "" {
		_ = os.Remove(s.socketPath)
	}
	if s.pidPath != "" {
		_ = os.Remove(s.pidPath)
	}
	return closeErr
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

type registerRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	body, err := readLimitedBody(r.Body, maxRequestBodySize)
	if err != nil {
		if errors.Is(err, errRequestBodyTooLarge) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	var req registerRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid register request", http.StatusBadRequest)
		return
	}

	agent, err := s.store.Register(req.Name)
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			http.Error(w, "agent already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, agent)
}

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	state, err := s.store.GetState(name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(state)
}

func (s *Server) handlePutState(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	body, err := readLimitedBody(r.Body, maxRequestBodySize)
	if err != nil {
		if errors.Is(err, errRequestBodyTooLarge) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if err := s.store.PutState(name, body); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, "agent not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// readLimitedBody は limit バイトまでのリクエストボディを読み取る。
func readLimitedBody(r io.Reader, limit int64) ([]byte, error) {
	limited := io.LimitReader(r, limit+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, errRequestBodyTooLarge
	}
	return body, nil
}

func removeStaleSocket(socketPath string) error {
	if _, err := os.Stat(socketPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat socket path: %w", err)
	}

	conn, err := net.Dial("unix", socketPath)
	if err == nil {
		_ = conn.Close()
		return fmt.Errorf("daemon is already running")
	}

	if err := os.Remove(socketPath); err != nil {
		return fmt.Errorf("remove stale socket: %w", err)
	}

	return nil
}

// SignalProcess は PID ファイルに記録された daemon にシグナルを送る。
func SignalProcess(sig syscall.Signal) error {
	pidPath, err := runtime.PIDPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(pidPath)
	if err != nil {
		return err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(sig)
}
