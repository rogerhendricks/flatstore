package flathub

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
)

func TestParseProgressStream(t *testing.T) {
	var capturedEvent string
	var capturedPercent int

	// Dummy emitter to catch what the parser extracts
	mockEmitter := func(name string, data ...interface{}) {
		capturedEvent = name
		if len(data) > 0 {
			payload := data[0].(ProgressPayload)
			capturedPercent = payload.Percentage
		}
	}

	mgr := NewSystemManager(mockEmitter)

	// Fake terminal data outputted by Flatpak installer sequence
	var mockTerminalOutput bytes.Buffer
	mockTerminalOutput.WriteString("Installing org.gimp.GIMP/x86_64/stable\n")
	mockTerminalOutput.WriteString("Downloading: [==        ]  23%  12.4MB/s\n")
	mockTerminalOutput.WriteString("Downloading: [=====     ]  57%  14.1MB/s\n")

	// Convert buffer to ReadCloser type to satisfy interface requirements
	readCloser := io.NopCloser(&mockTerminalOutput)

	// Execute internal reader loop manually
	mgr.parseProgressStream(readCloser, "org.gimp.GIMP")

	if capturedEvent != "flatpak:progress" {
		t.Errorf("Expected event 'flatpak:progress', got '%s'", capturedEvent)
	}

	// The last line passed was 57%
	if capturedPercent != 57 {
		t.Errorf("Expected progress extraction to read final calculation of 57, got %d", capturedPercent)
	}
}

// fakeExecCommand intercepts the call and runs the test binary instead of "flatpak"
func fakeExecCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	// Set our secret environment variable
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// TestHelperProcess isn't a real test. It acts as our fake "flatpak" binary.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Read the arguments passed to our fake flatpak
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	// args[0] is "flatpak"
	// args[1] is the action: "install", "uninstall", or "update"
	if len(args) > 1 {
		action := args[1]
		// Simulate standard flatpak output
		fmt.Printf("Starting mock %s...\n", action)
		fmt.Println("Downloading: [==        ]  50%  12.4MB/s")
		fmt.Println("Downloading: [==========] 100%  14.1MB/s")
	}

	os.Exit(0)
}

func TestSystemManager_Actions(t *testing.T) {
	// 1. Swap the real exec command with our fake one
	originalExec := execCommandContext
	execCommandContext = fakeExecCommand
	defer func() { execCommandContext = originalExec }() // Ensure we put it back!

	var capturedStatus string
	var capturedPercent int

	// 2. Setup our mock Wails event emitter
	mockEmitter := func(name string, data ...interface{}) {
		if len(data) > 0 {
			payload, ok := data[0].(ProgressPayload)
			if ok {
				capturedStatus = payload.Status
				capturedPercent = payload.Percentage
			}
		}
	}

	mgr := NewSystemManager(mockEmitter)
	ctx := context.Background()
	appID := "org.gimp.GIMP"

	// --- TEST INSTALL ---
	t.Run("Install", func(t *testing.T) {
		err := mgr.ExecuteWithProgress(ctx, appID, "install", "flathub", appID, "--user")
		if err != nil {
			t.Fatalf("Install expected no error, got: %v", err)
		}
		if capturedStatus != "completed" || capturedPercent != 100 {
			t.Errorf("Install expected 100%% completion, got status: %s, percent: %d", capturedStatus, capturedPercent)
		}
	})

	// --- TEST UPDATE ---
	t.Run("Update", func(t *testing.T) {
		err := mgr.ExecuteWithProgress(ctx, appID, "update", appID, "--user")
		if err != nil {
			t.Fatalf("Update expected no error, got: %v", err)
		}
		if capturedStatus != "completed" {
			t.Errorf("Update expected completion, got %s", capturedStatus)
		}
	})

	// --- TEST UNINSTALL ---
	t.Run("Uninstall", func(t *testing.T) {
		err := mgr.ExecuteWithProgress(ctx, appID, "uninstall", appID, "--user")
		if err != nil {
			t.Fatalf("Uninstall expected no error, got: %v", err)
		}
	})
}
