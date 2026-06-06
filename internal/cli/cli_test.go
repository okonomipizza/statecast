package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/okonomipizza/statecast/internal/client"
	"github.com/okonomipizza/statecast/internal/daemon"
	"github.com/okonomipizza/statecast/internal/store"
)

func startTestDaemon(t *testing.T) {
	t.Helper()

	// macOS の Unix ソケットパス長制限（104 バイト）を避けるため /tmp 配下に置く
	runtimeDir, err := os.MkdirTemp("/tmp", "sc")
	if err != nil {
		t.Fatalf("テスト用ランタイムディレクトリの作成に失敗した: %v", err)
	}
	t.Setenv("STATECAST_RUNTIME_DIR", runtimeDir)

	server := daemon.NewServer(store.New())
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	t.Cleanup(func() {
		_ = server.Close()
		_ = os.RemoveAll(runtimeDir)
	})

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		c, err := client.New()
		if err == nil && c.Ping() == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("daemon の起動待ちがタイムアウトした")
}

func registerAgent(t *testing.T, name string) {
	t.Helper()

	c, err := client.New()
	if err != nil {
		t.Fatalf("クライアント作成に失敗した: %v", err)
	}
	if _, err := c.Register(name); err != nil {
		t.Fatalf("register に失敗した: %v", err)
	}
}

func TestClientWhenDaemonNotRunning(t *testing.T) {
	// daemon 未起動時はわかりやすいエラーになること
	t.Setenv("STATECAST_RUNTIME_DIR", t.TempDir())

	c, err := client.New()
	if err != nil {
		t.Fatalf("クライアント作成に失敗した: %v", err)
	}

	if err := c.Ping(); err == nil || !strings.Contains(err.Error(), "daemon is not running") {
		t.Fatalf("期待したエラーが返らなかった: %v", err)
	}
}

func TestRegisterUpdateGetListFlow(t *testing.T) {
	// register / update / get / list が連携して動作すること
	startTestDaemon(t)
	registerAgent(t, "Agent A")

	c, err := client.New()
	if err != nil {
		t.Fatalf("クライアント作成に失敗した: %v", err)
	}

	payload := json.RawMessage(`{"key":"value"}`)
	if err := c.UpdateState("Agent A", payload); err != nil {
		t.Fatalf("update に失敗した: %v", err)
	}

	got, err := c.GetState("Agent A")
	if err != nil {
		t.Fatalf("get に失敗した: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("get の結果が一致しない: got %s", got)
	}

	agents, err := c.List()
	if err != nil {
		t.Fatalf("list に失敗した: %v", err)
	}
	if len(agents) != 1 || agents[0].Name != "Agent A" {
		t.Fatalf("list の結果が一致しない: %+v", agents)
	}
}

func TestUpdateRequiresRegisteredAgent(t *testing.T) {
	// 未登録エージェントへの update は拒否されること
	startTestDaemon(t)

	c, err := client.New()
	if err != nil {
		t.Fatalf("クライアント作成に失敗した: %v", err)
	}

	if err := c.UpdateState("missing", json.RawMessage(`{"key":"value"}`)); err == nil {
		t.Fatal("未登録エージェントへの update が成功してしまった")
	}
}

func TestGetCommandRequiresName(t *testing.T) {
	// get は --name 未指定時にエラーになること
	startTestDaemon(t)

	err := getCmd.RunE(getCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "agent name is required") {
		t.Fatalf("期待したエラーが返らなかった: %v", err)
	}
}

func TestListCommandOutput(t *testing.T) {
	// list コマンドが登録済みエージェントを表示すること
	startTestDaemon(t)
	registerAgent(t, "Agent A")

	var buf bytes.Buffer
	listCmd.SetOut(&buf)
	listCmd.SetErr(&buf)

	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatalf("list コマンドの実行に失敗した: %v", err)
	}

	output := buf.String()
	if output != "# name\nAgent A\n" {
		t.Fatalf("list の出力が一致しない: %q", output)
	}
}

func TestGetCommandWithName(t *testing.T) {
	// get --name で状態を取得できること
	startTestDaemon(t)
	registerAgent(t, "Agent B")

	c, err := client.New()
	if err != nil {
		t.Fatalf("クライアント作成に失敗した: %v", err)
	}
	if err := c.UpdateState("Agent B", json.RawMessage(`{"from":"b"}`)); err != nil {
		t.Fatalf("update に失敗した: %v", err)
	}

	var buf bytes.Buffer
	getCmd.SetOut(&buf)
	getCmd.SetErr(&buf)
	if err := getCmd.Flags().Set("name", "Agent B"); err != nil {
		t.Fatalf("name フラグの設定に失敗した: %v", err)
	}

	if err := getCmd.RunE(getCmd, nil); err != nil {
		t.Fatalf("get コマンドの実行に失敗した: %v", err)
	}

	if got := buf.String(); got != `{"from":"b"}` {
		t.Fatalf("get の出力が一致しない: %q", got)
	}
}

func TestUpdateCommandWithName(t *testing.T) {
	// update --name で状態を更新できること
	startTestDaemon(t)
	registerAgent(t, "Agent Env")

	if err := updateCmd.Flags().Set("name", "Agent Env"); err != nil {
		t.Fatalf("name フラグの設定に失敗した: %v", err)
	}
	if err := updateCmd.RunE(updateCmd, []string{`{"via":"name-flag"}`}); err != nil {
		t.Fatalf("update コマンドの実行に失敗した: %v", err)
	}

	c, err := client.New()
	if err != nil {
		t.Fatalf("クライアント作成に失敗した: %v", err)
	}

	state, err := c.GetState("Agent Env")
	if err != nil {
		t.Fatalf("get に失敗した: %v", err)
	}
	if string(state) != `{"via":"name-flag"}` {
		t.Fatalf("保存された状態が一致しない: %s", state)
	}
}

func TestRegisterCommandOutput(t *testing.T) {
	// register は登録した name を出力すること
	startTestDaemon(t)

	if err := registerCmd.Flags().Set("name", "Agent A"); err != nil {
		t.Fatalf("name フラグの設定に失敗した: %v", err)
	}

	var buf bytes.Buffer
	registerCmd.SetOut(&buf)
	registerCmd.SetErr(&buf)

	if err := registerCmd.RunE(registerCmd, nil); err != nil {
		t.Fatalf("register コマンドの実行に失敗した: %v", err)
	}

	if strings.TrimSpace(buf.String()) != "Agent A" {
		t.Fatalf("register の出力が一致しない: %q", buf.String())
	}
}

func TestClientAgentNameEdgeCases(t *testing.T) {
	// 特殊文字を含む name でも register / update / get が動作すること
	cases := []struct {
		name string
	}{
		{name: "Agent A"},
		{name: "エージェント"},
		{name: "agent/with/slash"},
		{name: "agent space"},
	}

	startTestDaemon(t)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := client.New()
			if err != nil {
				t.Fatalf("クライアント作成に失敗した: %v", err)
			}

			if _, err := c.Register(tc.name); err != nil {
				t.Fatalf("register に失敗した: %v", err)
			}

			payload := json.RawMessage(`{"ok":true}`)
			if err := c.UpdateState(tc.name, payload); err != nil {
				t.Fatalf("update に失敗した: %v", err)
			}

			got, err := c.GetState(tc.name)
			if err != nil {
				t.Fatalf("get に失敗した: %v", err)
			}
			if string(got) != string(payload) {
				t.Fatalf("state が一致しない: got %s", got)
			}
		})
	}
}

func TestStopCommandWhenNotRunning(t *testing.T) {
	// stop は daemon 未起動時にエラーになること
	runtimeDir, err := os.MkdirTemp("/tmp", "sc")
	if err != nil {
		t.Fatalf("テスト用ランタイムディレクトリの作成に失敗した: %v", err)
	}
	t.Setenv("STATECAST_RUNTIME_DIR", runtimeDir)

	err = stopCmd.RunE(stopCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "daemon is not running") {
		t.Fatalf("期待したエラーが返らなかった: %v", err)
	}
}

func TestMain(m *testing.M) {
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	os.Exit(m.Run())
}
