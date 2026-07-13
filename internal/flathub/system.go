package flathub

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Allow overriding the command executor for unit testing
var execCommandContext = exec.CommandContext

// ProgressPayload is the structured data we send to the frontend
type ProgressPayload struct {
	AppID      string `json:"appId"`
	Status     string `json:"status"`     // "downloading", "installing", "completed", "error"
	Percentage int    `json:"percentage"` // 0 to 100
}

type SystemManager struct {
	// Function pointer to Wails event emitter so this package remains independent of the main package
	emitEvent func(eventName string, optionalData ...interface{})
}

type InstalledApp struct {
	AppID           string `json:"appId"`
	Name            string `json:"name"`
	Version         string `json:"version"`
	UpdateAvailable bool   `json:"updateAvailable"`
}

func NewSystemManager(emitter func(string, ...interface{})) *SystemManager {
	return &SystemManager{
		emitEvent: emitter,
	}
}

// ExecuteWithProgress runs the flatpak command and pipes stdout to a scanner
func (m *SystemManager) ExecuteWithProgress(ctx context.Context, appID string, args ...string) error {
	// Merge mandatory operational flags
	defaultArgs := []string{"--noninteractive", "-y"}
	finalArgs := append(args, defaultArgs...)

	// cmd := exec.CommandContext(ctx, "flatpak", finalArgs...)
	cmd := execCommandContext(ctx, "flatpak", finalArgs...)

	// Create a pipe to capture standard output
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Direct stderr to a separate buffer so we don't pollute our progress parser with error traces
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start flatpak command: %w", err)
	}

	// Read from the pipe concurrently so we don't block execution
	go m.parseProgressStream(stdoutPipe, appID)

	// Wait for the process to terminate safely
	if err := cmd.Wait(); err != nil {
		m.emitEvent("flatpak:progress", ProgressPayload{
			AppID:      appID,
			Status:     "error",
			Percentage: 0,
		})
		return fmt.Errorf("flatpak command failed: %w, stderr: %s", err, stderrBuf.String())
	}

	// Send absolute confirmation to the UI layer
	m.emitEvent("flatpak:progress", ProgressPayload{
		AppID:      appID,
		Status:     "completed",
		Percentage: 100,
	})

	return nil
}

// Regex to capture standard flatpak percentage markers like " 45%" or "100%"
var percentRegex = regexp.MustCompile(`(\d+)%`)

func (m *SystemManager) parseProgressStream(reader io.ReadCloser, appID string) {
	defer reader.Close()
	scanner := bufio.NewScanner(reader)

	// Flatpak updates its terminal lines quickly.
	// Scanner reads token-by-token or line-by-line automatically.
	for scanner.Scan() {
		line := scanner.Text()

		// Look for percentage patterns in the string output
		matches := percentRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			percentVal, err := strconv.Atoi(matches[1])
			if err == nil {
				status := "downloading"
				if percentVal == 100 {
					status = "installing" // Downloading finished, moving to local deployment
				}

				// Emit real-time tracking payload to Wails
				m.emitEvent("flatpak:progress", ProgressPayload{
					AppID:      appID,
					Status:     status,
					Percentage: percentVal,
				})
			}
		}
	}
}

// ListInstalledApps queries the system for installed flatpaks and checks for updates
func (m *SystemManager) ListInstalledApps(ctx context.Context) ([]InstalledApp, error) {
	// 1. Get installed apps (tab-separated for easy parsing)
	cmd := execCommandContext(ctx, "flatpak", "list", "--app", "--columns=application,name,version")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}

	// 2. Get updates
	updateCmd := execCommandContext(ctx, "flatpak", "remote-ls", "--updates", "--app", "--columns=application")
	updateOut, _ := updateCmd.Output() // Ignore error, if it fails we just assume no updates

	// Parse updates into a map for fast O(1) lookup
	updates := make(map[string]bool)
	for _, line := range strings.Split(string(updateOut), "\n") {
		cleanLine := strings.TrimSpace(line)
		if cleanLine != "" {
			updates[cleanLine] = true
		}
	}

	// Parse the installed apps
	var apps []InstalledApp
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			appID := strings.TrimSpace(parts[0])
			apps = append(apps, InstalledApp{
				AppID:           appID,
				Name:            strings.TrimSpace(parts[1]),
				Version:         strings.TrimSpace(parts[2]),
				UpdateAvailable: updates[appID],
			})
		}
	}

	return apps, nil
}

func (m *SystemManager) RunApp(appID string) error {
	cmd := execCommandContext(context.Background(), "flatpak", "run", appID)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start application %s: %w", appID, err)
	}
	return nil
}

