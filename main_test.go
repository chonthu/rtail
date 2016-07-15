package main

import (
	"testing"
)

func TestParseServers(t *testing.T) {
	t.Parallel()

	servers := []string{"web[1-2].spinmedia.com"}
	out := parseServers(servers)

	if len(out) != 2 {
		t.Error("failed hoe")
	}
}

func TestInvalidParseServers(t *testing.T) {
	t.Parallel()

	servers := []string{"web[1-].spinmedia.com"}
	out := parseServers(servers)

	if len(out) != 0 {
		t.Error("failed hoe")
	}
}
