package main

import (
	"testing"
)

func TestParseServers(t *testing.T) {
	t.Parallel()

	servers := []Server{Server{host: "web[1-2].spinmedia.com"}}
	out, err := rangeSplitServers(servers)

	if len(out) != 2 {
		t.Error(err, "invalid number of servers found")
	}
}

func TestInvalidParseServers(t *testing.T) {
	t.Parallel()

	servers := []Server{Server{host: "web[1-].spinmedia.com"}}
	out, err := rangeSplitServers(servers)

	if len(out) != 0 {
		t.Error(err, "invalid number of errors found")
	}
}
