// Copyright (c) 2024-2026 Victor Gallardo Sanchez. All rights reserved.
// Licensed under the Business Source License 1.1.
// See the LICENSE file in the repository root for full license terms.

package legal

import (
	"encoding/json"
	"os"
	"time"
)

// EventType represents the mandatory event types from Art. 9.
type EventType string

const (
	// EventStart marks the beginning of a system operation.
	EventStart EventType = "START"
	// EventStop marks the end of a system operation.
	EventStop EventType = "STOP"
	// EventAnomaly signals a detected cryptographic or integrity anomaly.
	EventAnomaly EventType = "ANOMALY"
	// EventRestore marks a successful recovery from an anomaly.
	EventRestore EventType = "RESTORE"
	// EventExport records a data export event for audit purposes.
	EventExport EventType = "EXPORT"
)

// EventRecord represents a mandatory audit log entry.
type EventRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// LogEvent records a mandatory event to the audit trail.
func LogEvent(eType EventType, message string) error {
	record := EventRecord{
		Timestamp: time.Now().UTC(),
		Type:      eType,
		Message:   message,
		Source:    "FacturaE-Sidecar",
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// In a real implementation, this would go to a secure database.
	// For the POC, we append to a local audit file.
	f, err := os.OpenFile("audit_trail.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}
