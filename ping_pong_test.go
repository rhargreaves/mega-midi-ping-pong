package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
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

func TestPingPong(t *testing.T) {
	drv, inID, outID := setupTestMIDI(t)
	defer drv.Close()

	// Start the fake Mega Drive responder
	// App sends on Bus 1, receives on Bus 2
	// Responder listens on Bus 1, sends on Bus 2
	stopFake := startFakeMegaDriveResponder(t, drv, outID, inID)
	defer close(stopFake)

	// Run 5 ping-pongs for the test
	times, durations := pingPong(drv, inID, outID, 5)

	if len(times) != 5 || len(durations) != 5 {
		t.Fatalf("Expected 5 ping-pongs, got %d", len(times))
	}

	for i, duration := range durations {
		if duration <= 0 {
			t.Errorf("Invalid duration at index %d: %v", i, duration)
		}
		if duration > 1000 {
			t.Errorf("Suspiciously long duration at index %d: %v", i, duration)
		}
	}
}

func TestListDevices(t *testing.T) {
	drv, err := rtmididrv.New()
	if err != nil {
		t.Fatalf("Failed to create MIDI driver: %v", err)
	}
	defer drv.Close()

	ins, err := drv.Ins()
	if err != nil {
		t.Fatalf("Failed to get input devices: %v", err)
	}

	outs, err := drv.Outs()
	if err != nil {
		t.Fatalf("Failed to get output devices: %v", err)
	}

	if len(ins) == 0 {
		t.Error("No input devices found")
	}

	if len(outs) == 0 {
		t.Error("No output devices found")
	}
}

// setupTestMIDI creates virtual MIDI devices for testing
func setupTestMIDI(t *testing.T) (*rtmididrv.Driver, int, int) {
	drv, err := rtmididrv.New()
	if err != nil {
		t.Fatalf("Failed to create MIDI driver: %v", err)
	}

	// In CI, we'll use virtual devices
	if os.Getenv("CI") != "" {
		return setupVirtualMIDI(t, drv)
	}

	// On macOS, we'll use IAC Driver
	return setupIACMIDI(t, drv)
}

func setupVirtualMIDI(t *testing.T, drv *rtmididrv.Driver) (*rtmididrv.Driver, int, int) {
	ins, err := drv.Ins()
	if err != nil {
		t.Fatalf("Failed to get input devices: %v", err)
	}

	outs, err := drv.Outs()
	if err != nil {
		t.Fatalf("Failed to get output devices: %v", err)
	}

	// Find virtual MIDI devices
	var inID, outID int = -1, -1
	for i, in := range ins {
		if in.String() == "Virtual MIDI" {
			inID = i
			break
		}
	}

	for i, out := range outs {
		if out.String() == "Virtual MIDI" {
			outID = i
			break
		}
	}

	if inID == -1 || outID == -1 {
		t.Fatal("Virtual MIDI devices not found")
	}

	return drv, inID, outID
}

func setupIACMIDI(t *testing.T, drv *rtmididrv.Driver) (*rtmididrv.Driver, int, int) {
	ins, err := drv.Ins()
	if err != nil {
		t.Fatalf("Failed to get input devices: %v", err)
	}

	outs, err := drv.Outs()
	if err != nil {
		t.Fatalf("Failed to get output devices: %v", err)
	}

	// Find IAC Driver devices
	// App sends on Bus 1, receives on Bus 2
	var inID, outID int = -1, -1
	for i, in := range ins {
		if in.String() == "IAC Driver Bus 2" {
			inID = i
			break
		}
	}

	for i, out := range outs {
		if out.String() == "IAC Driver Bus 1" {
			outID = i
			break
		}
	}

	if inID == -1 || outID == -1 {
		t.Fatal("IAC Driver devices not found. Please enable IAC Driver in Audio MIDI Setup")
	}

	return drv, inID, outID
}

// startFakeMegaDriveResponder listens for ping SysEx and responds with pong SysEx
func startFakeMegaDriveResponder(t *testing.T, drv *rtmididrv.Driver, listenID, sendID int) chan struct{} {
	ins, err := drv.Ins()
	if err != nil {
		t.Fatalf("Failed to get input devices: %v", err)
	}
	outs, err := drv.Outs()
	if err != nil {
		t.Fatalf("Failed to get output devices: %v", err)
	}

	// Find Bus 1 for listening and Bus 2 for sending
	var listenPort, sendPort int = -1, -1
	for i, in := range ins {
		if in.String() == "IAC Driver Bus 1" {
			listenPort = i
			break
		}
	}
	for i, out := range outs {
		if out.String() == "IAC Driver Bus 2" {
			sendPort = i
			break
		}
	}

	if listenPort == -1 || sendPort == -1 {
		t.Fatalf("Could not find required IAC Driver buses")
	}

	in := ins[listenPort]
	out := outs[sendPort]

	if err := in.Open(); err != nil {
		t.Fatalf("Failed to open fake input: %v", err)
	}
	if err := out.Open(); err != nil {
		in.Close()
		t.Fatalf("Failed to open fake output: %v", err)
	}

	stop := make(chan struct{})
	go func() {
		defer in.Close()
		defer out.Close()
		pingSysEx := []byte{0x00, 0x22, 0x77, 0x01}
		pongSysEx := []byte{0xF0, 0x00, 0x22, 0x77, 0x02, 0xF7}
		r := reader.New(
			reader.NoLogger(),
			reader.SysEx(func(_ *reader.Position, b []byte) {
				if bytes.Equal(b, pingSysEx) {
					// Respond with pong
					out.Write(pongSysEx)
				}
			}),
		)
		r.ListenTo(in)
		<-stop
	}()
	return stop
}
