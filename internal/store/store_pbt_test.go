package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"pgregory.net/rapid"
)

// drawValidJSON は rapid で有効な JSON バイト列を生成する。
func drawValidJSON(t *rapid.T) []byte {
	t.Helper()

	switch rapid.IntRange(0, 4).Draw(t, "kind") {
	case 0:
		return []byte("null")
	case 1:
		v, err := json.Marshal(rapid.Bool().Draw(t, "bool"))
		if err != nil {
			t.Fatalf("marshal bool: %v", err)
		}
		return v
	case 2:
		v, err := json.Marshal(rapid.Int64().Draw(t, "int"))
		if err != nil {
			t.Fatalf("marshal int: %v", err)
		}
		return v
	case 3:
		v, err := json.Marshal(rapid.String().Draw(t, "string"))
		if err != nil {
			t.Fatalf("marshal string: %v", err)
		}
		return v
	default:
		v, err := json.Marshal(map[string]int{"n": rapid.IntRange(-100, 100).Draw(t, "mapval")})
		if err != nil {
			t.Fatalf("marshal map: %v", err)
		}
		return v
	}
}

func TestPBTRegisterThenGetStateIsEmptyObject(t *testing.T) {
	// 登録直後の state は常に空オブジェクトであること
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.String().Filter(func(s string) bool { return s != "" }).Draw(t, "name")

		st := New()
		agent, err := st.Register(name)
		if err != nil {
			t.Fatalf("register failed: %v", err)
		}
		if agent.Name != name {
			t.Fatalf("name mismatch: got %q want %q", agent.Name, name)
		}
		if string(agent.State) != `{}` {
			t.Fatalf("initial state is not empty object: %q", agent.State)
		}

		state, err := st.GetState(name)
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}
		if string(state) != `{}` {
			t.Fatalf("GetState after register: got %q", state)
		}
	})
}

func TestPBTPutStateRoundtrip(t *testing.T) {
	// 有効な JSON は PutState → GetState で完全に復元されること
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.String().Filter(func(s string) bool { return s != "" }).Draw(t, "name")
		state := drawValidJSON(t)

		st := New()
		if _, err := st.Register(name); err != nil {
			t.Fatalf("register failed: %v", err)
		}

		raw := json.RawMessage(state)
		if err := st.PutState(name, raw); err != nil {
			t.Fatalf("PutState failed: %v", err)
		}

		got, err := st.GetState(name)
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}
		if !bytes.Equal(got, raw) {
			t.Fatalf("state mismatch: got %q want %q", got, raw)
		}
	})
}

func TestPBTInvalidJSONDoesNotMutateState(t *testing.T) {
	// 無効な JSON を PutState しても state は変化しないこと
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.String().Filter(func(s string) bool { return s != "" }).Draw(t, "name")
		valid := drawValidJSON(t)
		invalid := rapid.String().Filter(func(s string) bool { return !json.Valid([]byte(s)) }).Draw(t, "invalid")

		st := New()
		if _, err := st.Register(name); err != nil {
			t.Fatalf("register failed: %v", err)
		}
		if err := st.PutState(name, json.RawMessage(valid)); err != nil {
			t.Fatalf("PutState valid failed: %v", err)
		}

		before, err := st.GetState(name)
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}

		if err := st.PutState(name, json.RawMessage(invalid)); err == nil {
			t.Fatal("PutState accepted invalid JSON")
		}

		after, err := st.GetState(name)
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}
		if !bytes.Equal(before, after) {
			t.Fatalf("state mutated: before %q after %q", before, after)
		}
	})
}

func TestPBTListContainsExactlyRegisteredNames(t *testing.T) {
	// List は登録済み name を過不足なく返すこと
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(0, 20).Draw(t, "n")
		names := make([]string, n)
		for i := range names {
			names[i] = rapid.String().Filter(func(s string) bool { return s != "" }).Draw(t, "name")
		}

		st := New()
		expected := map[string]bool{}
		for _, name := range names {
			if expected[name] {
				if _, err := st.Register(name); !errors.Is(err, ErrAlreadyExists) {
					t.Fatalf("expected ErrAlreadyExists for duplicate %q", name)
				}
				continue
			}
			if _, err := st.Register(name); err != nil {
				t.Fatalf("register failed for %q: %v", name, err)
			}
			expected[name] = true
		}

		listed := st.List()
		if len(listed) != len(expected) {
			t.Fatalf("list count mismatch: got %d want %d", len(listed), len(expected))
		}
		for _, agent := range listed {
			if !expected[agent.Name] {
				t.Fatalf("unexpected agent in list: %q", agent.Name)
			}
		}
	})
}

func TestPBTGetStateReturnsIndependentCopy(t *testing.T) {
	// GetState の戻り値を変更しても store 内の state は変わらないこと
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.String().Filter(func(s string) bool { return s != "" }).Draw(t, "name")
		state := drawValidJSON(t)

		st := New()
		if _, err := st.Register(name); err != nil {
			t.Fatalf("register failed: %v", err)
		}
		if err := st.PutState(name, json.RawMessage(state)); err != nil {
			t.Fatalf("PutState failed: %v", err)
		}

		got, err := st.GetState(name)
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}
		if len(got) > 0 {
			got[0] ^= 0xff
		}

		again, err := st.GetState(name)
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}
		if !bytes.Equal(again, json.RawMessage(state)) {
			t.Fatalf("store state was mutated via returned slice")
		}
	})
}
