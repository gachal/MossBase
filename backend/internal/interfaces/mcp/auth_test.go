package mcp

import (
	"context"
	"testing"
)

func TestNewMCPAuth_Disabled(t *testing.T) {
	auth := NewMCPAuth(nil, 1)
	if auth.enabled {
		t.Error("expected auth to be disabled with nil keys")
	}

	uid, err := auth.Authenticate("")
	if err != nil {
		t.Fatalf("expected no error when auth disabled, got: %v", err)
	}
	if uid != 1 {
		t.Errorf("expected default user ID 1, got %d", uid)
	}
}

func TestNewMCPAuth_EmptyKeys(t *testing.T) {
	auth := NewMCPAuth([]string{"", "  "}, 5)
	if auth.enabled {
		t.Error("expected auth to be disabled with only whitespace keys")
	}
}

func TestMCPAuth_Authenticate_ValidKey(t *testing.T) {
	auth := NewMCPAuth([]string{"key1", "key2"}, 10)
	if !auth.enabled {
		t.Fatal("expected auth to be enabled")
	}

	uid, err := auth.Authenticate("key1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if uid != 10 {
		t.Errorf("expected user ID 10, got %d", uid)
	}

	uid, err = auth.Authenticate("key2")
	if err != nil {
		t.Fatalf("expected no error for second key, got: %v", err)
	}
	if uid != 10 {
		t.Errorf("expected user ID 10, got %d", uid)
	}
}

func TestMCPAuth_Authenticate_InvalidKey(t *testing.T) {
	auth := NewMCPAuth([]string{"valid-key"}, 1)

	_, err := auth.Authenticate("wrong-key")
	if err == nil {
		t.Error("expected error for invalid key")
	}

	_, err = auth.Authenticate("")
	if err == nil {
		t.Error("expected error for empty key when auth enabled")
	}
}

func TestWithUserID_GetUserID(t *testing.T) {
	ctx := WithUserID(context.Background(), 42)

	uid := GetUserID(ctx, 1)
	if uid != 42 {
		t.Errorf("expected user ID 42 from context, got %d", uid)
	}
}

func TestGetUserID_Default(t *testing.T) {
	uid := GetUserID(context.Background(), 99)
	if uid != 99 {
		t.Errorf("expected default user ID 99 when not in context, got %d", uid)
	}
}
