package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

func TestUnmarshalPath(t *testing.T) {
	t.Run("UnmarshalPath", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
			t.Run("success", func(t *testing.T) {
				var v int
				if err := json.UnmarshalPath("a.b", src, &v); err != nil {
					t.Fatal(err)
				}
				if v != 10 {
					t.Fatal("failed to unmarshal path")
				}
			})
			t.Run("failure", func(t *testing.T) {
				var v int
				if err := json.UnmarshalPath("a.c", src, &v); err == nil {
					t.Fatal("expected error")
				}
			})
		})
		t.Run("bool", func(t *testing.T) {
			src := []byte(`{"a":{"b":10,"c":true},"b":"text"}`)
			t.Run("success", func(t *testing.T) {
				var v bool
				if err := json.UnmarshalPath("a.c", src, &v); err != nil {
					t.Fatal(err)
				}
				if !v {
					t.Fatal("failed to unmarshal path")
				}
			})
			t.Run("failure", func(t *testing.T) {
				var v bool
				if err := json.UnmarshalPath("a.b", src, &v); err == nil {
					t.Fatal("expected error")
				}
			})
		})
	})
}
