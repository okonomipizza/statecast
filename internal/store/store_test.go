package store

import (
	"encoding/json"
	"testing"
)

func TestRegister(t *testing.T) {
	// name で登録できること
	st := New()

	agent, err := st.Register("Agent A")
	if err != nil {
		t.Fatalf("register に失敗した: %v", err)
	}
	if agent.Name != "Agent A" {
		t.Fatalf("name が一致しない: %s", agent.Name)
	}
	if string(agent.State) != `{}` {
		t.Fatalf("初期 state が空オブジェクトではない: %s", agent.State)
	}
}

func TestRegisterRejectsDuplicateName(t *testing.T) {
	// 同名の登録は拒否されること
	st := New()

	if _, err := st.Register("Agent A"); err != nil {
		t.Fatalf("1 回目の register に失敗した: %v", err)
	}
	if _, err := st.Register("Agent A"); err != ErrAlreadyExists {
		t.Fatalf("期待した ErrAlreadyExists が返らなかった: %v", err)
	}
}

func TestPutStateRequiresRegistration(t *testing.T) {
	// 未登録エージェントへの update は拒否されること
	st := New()

	if err := st.PutState("missing", json.RawMessage(`{"key":"value"}`)); err != ErrNotFound {
		t.Fatalf("期待した ErrNotFound が返らなかった: %v", err)
	}
}

func TestRegisterPutAndGetState(t *testing.T) {
	// register 後に state を更新・取得できること
	st := New()

	if _, err := st.Register("Agent A"); err != nil {
		t.Fatalf("register に失敗した: %v", err)
	}

	if err := st.PutState("Agent A", json.RawMessage(`{"task":"test"}`)); err != nil {
		t.Fatalf("state の保存に失敗した: %v", err)
	}

	state, err := st.GetState("Agent A")
	if err != nil {
		t.Fatalf("state の取得に失敗した: %v", err)
	}
	if string(state) != `{"task":"test"}` {
		t.Fatalf("取得した state が一致しない: got %s", state)
	}
}

func TestPutStateRejectsInvalidJSON(t *testing.T) {
	// 不正な JSON は拒否されること
	st := New()

	if _, err := st.Register("Agent A"); err != nil {
		t.Fatalf("register に失敗した: %v", err)
	}
	if err := st.PutState("Agent A", json.RawMessage(`{invalid`)); err == nil {
		t.Fatal("不正な JSON が拒否されなかった")
	}
}

func TestList(t *testing.T) {
	// List が name 順で全エージェントを返すこと
	st := New()

	if _, err := st.Register("Agent B"); err != nil {
		t.Fatalf("Agent B の register に失敗した: %v", err)
	}
	if _, err := st.Register("Agent A"); err != nil {
		t.Fatalf("Agent A の register に失敗した: %v", err)
	}

	entries := st.List()
	if len(entries) != 2 {
		t.Fatalf("一覧件数が一致しない: got %d", len(entries))
	}
	if entries[0].Name != "Agent A" || entries[1].Name != "Agent B" {
		t.Fatalf("一覧の並び順が一致しない: %+v", entries)
	}
}

func TestGetStateNotFound(t *testing.T) {
	// 未登録エージェントは ErrNotFound になること
	st := New()

	if _, err := st.GetState("missing"); err != ErrNotFound {
		t.Fatalf("期待した ErrNotFound が返らなかった: %v", err)
	}
}
