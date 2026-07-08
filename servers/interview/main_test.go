package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestSanitizeDatabaseURLRedactsEscapedPassword(t *testing.T) {
	password := "p@ss/w#rd"
	t.Setenv("DB_PASSWORD", password)

	escaped := strings.TrimPrefix(url.UserPassword("", password).String(), ":")
	dsn := "postgres://user:" + escaped + "@localhost:5432/db"

	sanitized := sanitizeDatabaseURL(dsn)

	if strings.Contains(sanitized, password) {
		t.Fatalf("raw password leaked: %s", sanitized)
	}
	if strings.Contains(sanitized, escaped) {
		t.Fatalf("escaped password leaked: %s", sanitized)
	}
}
