package flathub

import (
	"bytes"
	"io"
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
