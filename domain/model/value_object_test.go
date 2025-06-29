package model

import (
	"testing"
)

func TestNewId(t *testing.T) {
	_, err := NewId(0)
	if err == nil {
		t.Error("expected error for id = 0")
	}
	_, err = NewId(-5)
	if err == nil {
		t.Error("expected error for negative id")
	}
	id, err := NewId(10)
	if err != nil || id != 10 {
		t.Errorf("expected id=10, got %v, err=%v", id, err)
	}
}

func TestNewVersion(t *testing.T) {
	_, err := NewVersion(-1)
	if err == nil {
		t.Error("expected error for negative version")
	}
	v, err := NewVersion(0)
	if err != nil || v != 0 {
		t.Errorf("expected version=0, got %v, err=%v", v, err)
	}
	v, err = NewVersion(5)
	if err != nil || v != 5 {
		t.Errorf("expected version=5, got %v, err=%v", v, err)
	}
}

func TestNewTitle(t *testing.T) {
	_, err := NewTitle("")
	if err == nil {
		t.Error("expected error for empty title")
	}
	long := ""
	for i := 0; i < MaxTitleLength+1; i++ {
		long += "a"
	}
	_, err = NewTitle(long)
	if err == nil {
		t.Error("expected error for too long title")
	}
	title, err := NewTitle("Valid Title")
	if err != nil || title != "Valid Title" {
		t.Errorf("expected title='Valid Title', got %v, err=%v", title, err)
	}
}

func TestNewISBN(t *testing.T) {
	_, err := NewISBN("")
	if err == nil {
		t.Error("expected error for empty ISBN")
	}
	long := ""
	for i := 0; i < MaxISBNLength+1; i++ {
		long += "1"
	}
	_, err = NewISBN(long)
	if err == nil {
		t.Error("expected error for too long ISBN")
	}
	isbn, err := NewISBN("1234567890")
	if err != nil || isbn != "1234567890" {
		t.Errorf("expected isbn='1234567890', got %v, err=%v", isbn, err)
	}
}

func TestNewURL(t *testing.T) {
	_, err := NewURL("")
	if err == nil {
		t.Error("expected error for empty URL")
	}
	_, err = NewURL("not-a-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
	url, err := NewURL("https://example.com")
	if err != nil || url != "https://example.com" {
		t.Errorf("expected url='https://example.com', got %v, err=%v", url, err)
	}
	url, err = NewURL("http://example.com/path")
	if err != nil || url != "http://example.com/path" {
		t.Errorf("expected url='http://example.com/path', got %v, err=%v", url, err)
	}
}
