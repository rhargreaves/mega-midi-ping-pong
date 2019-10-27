package main

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestMain(t *testing.T) {

	cmd := exec.Command("go", "run", ".", "-list")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Expected to succeed, but failed: %s %s", err, out.String())
	}
}
