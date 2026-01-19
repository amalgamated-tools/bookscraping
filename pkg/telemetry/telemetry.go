package telemetry

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
)

const installIDPath = "/data/install_id"

type Payload struct {
	InstallID string `json:"install_id"`
	Version   string `json:"version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Timestamp string `json:"timestamp"`
}

func Send(version string) {
	// Opt-out
	if os.Getenv("TELEMETRY_ENABLED") == "false" {
		slog.Info("Telemetry disabled via TELEMETRY_ENABLED=false")
		return
	}

	endpoint := os.Getenv("TELEMETRY_ENDPOINT")
	if endpoint == "" {
		slog.Info("Telemetry endpoint not set, skipping telemetry")
		return
	}

	// Only send once per install
	if _, err := os.Stat(installIDPath); err == nil {
		slog.Info("Telemetry already sent for this install, skipping")
		return
	}

	id := uuid.New().String()
	err := os.WriteFile(installIDPath, []byte(id), 0644)
	if err != nil {
		slog.Error("Failed to write install ID", "error", err)
		return
	}

	payload := Payload{
		InstallID: id,
		Version:   version,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("Failed to create telemetry request", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to send telemetry request", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Telemetry request failed", "status", resp.StatusCode)
		return
	}

	// write out response to log
	slog.Info("Telemetry sent successfully")
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read telemetry response", "error", err)
		return
	}
	slog.Info("Telemetry response", "body", string(body))
}
