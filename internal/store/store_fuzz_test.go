package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func FuzzPutStateValidJSON(f *testing.F) {
	// 有効な JSON は PutState 後に GetState で同一内容が返ること
	f.Add([]byte(`{"key":"value"}`))
	f.Add([]byte(`[1,2,3]`))
	f.Add([]byte(`"string"`))
	f.Add([]byte(`null`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if !json.Valid(data) {
			return
		}

		st := New()
		if _, err := st.Register("agent"); err != nil {
			t.Fatalf("register failed: %v", err)
		}

		if err := st.PutState("agent", data); err != nil {
			t.Fatalf("PutState failed for valid JSON: %v", err)
		}

		got, err := st.GetState("agent")
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}
		if !bytes.Equal(got, data) {
			t.Fatalf("state mismatch: got %q want %q", got, data)
		}
	})
}

func FuzzPutStateInvalidJSON(f *testing.F) {
	// 無効な JSON は拒否され、既存 state を変更しないこと
	f.Add([]byte(`{invalid`))
	f.Add([]byte(`{"key":}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if json.Valid(data) {
			return
		}

		st := New()
		if _, err := st.Register("agent"); err != nil {
			t.Fatalf("register failed: %v", err)
		}

		before, err := st.GetState("agent")
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}

		if err := st.PutState("agent", data); err == nil {
			t.Fatal("PutState accepted invalid JSON")
		}

		after, err := st.GetState("agent")
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}
		if !bytes.Equal(before, after) {
			t.Fatalf("state changed after rejected PutState: before %q after %q", before, after)
		}
	})
}

func FuzzRegisterName(f *testing.F) {
	// 空 name は拒否、非空 name は登録可、同名の再登録は ErrAlreadyExists になること
	f.Add("agent-a")
	f.Add("")
	f.Add("エージェント")

	f.Fuzz(func(t *testing.T, name string) {
		st := New()

		_, err := st.Register(name)
		if name == "" {
			if err == nil {
				t.Fatal("empty name was accepted")
			}
			return
		}
		if err != nil {
			t.Fatalf("first Register failed: %v", err)
		}

		_, err = st.Register(name)
		if !errors.Is(err, ErrAlreadyExists) {
			t.Fatalf("expected ErrAlreadyExists, got %v", err)
		}
	})
}
