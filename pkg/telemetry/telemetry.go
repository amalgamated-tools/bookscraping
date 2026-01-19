package telemetry

import (
	"bytes"
	"encoding/json"
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
		return
	}

	endpoint := os.Getenv("TELEMETRY_ENDPOINT")
	if endpoint == "" {
		return
	}

	// Only send once per install
	if _, err := os.Stat(installIDPath); err == nil {
		return
	}

	id := uuid.New().String()
	_ = os.WriteFile(installIDPath, []byte(id), 0644)

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
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	_, _ = client.Do(req)
}
