package main

import (
	"testing"

	"github.com/rendon/testcli"
)

func TestMain(t *testing.T) {
	testcli.Run("./mega-midi-ping-pong")
	if !testcli.Success() {
		t.Fatalf("Expected to succeed, but failed: %s", testcli.Error())
	}
}
