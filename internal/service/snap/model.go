package snap

import (
	"time"
)

type SnapshotData struct {
	Date time.Time `json:"date,omitempty"`
	ID   string    `json:"id,omitempty"`
}

type Request struct {
	SnapID string `json:"id,omitempty"`
}

type GetHistoryResponse struct {
	Snapshots []SnapshotData `json:"snapshots,omitempty"`
}

type RestoreSnapResponse struct {
	Date   time.Time `json:"date,omitempty"`
	ID     string    `json:"id,omitempty"`
	Status string    `json:"status,omitempty"`
}
