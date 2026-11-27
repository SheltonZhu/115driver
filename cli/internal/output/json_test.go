package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SheltonZhu/115driver/pkg/driver"
)

type fileJSONInput interface {
	GetName() string
	GetSize() int64
	IsDir() bool
	GetID() string
	ModTime() time.Time
}

type otherJSONFile struct{}

func (otherJSONFile) GetName() string    { return "other" }
func (otherJSONFile) GetSize() int64     { return 0 }
func (otherJSONFile) IsDir() bool        { return false }
func (otherJSONFile) GetID() string      { return "other-id" }
func (otherJSONFile) ModTime() time.Time { return time.Time{} }

func TestFileToJSONParentIDContract(t *testing.T) {
	tests := []struct {
		name         string
		file         fileJSONInput
		wantParentID string
		wantKey      bool
	}{
		{
			name:         "driver file with parent",
			file:         &driver.File{ParentID: "directory-123"},
			wantParentID: "directory-123",
			wantKey:      true,
		},
		{
			name:         "driver file in root",
			file:         &driver.File{ParentID: "0"},
			wantParentID: "0",
			wantKey:      true,
		},
		{
			name:    "driver file with unknown parent",
			file:    &driver.File{},
			wantKey: false,
		},
		{
			name:    "non-driver file",
			file:    otherJSONFile{},
			wantKey: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(FileToJSON(tt.file))
			if err != nil {
				t.Fatalf("marshal JSON file: %v", err)
			}

			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatalf("unmarshal JSON file: %v", err)
			}

			rawParentID, ok := fields["parent_id"]
			if !tt.wantKey {
				if ok {
					t.Fatalf("parent_id = %s, want key omitted", rawParentID)
				}
				return
			}
			if !ok {
				t.Fatal("parent_id key is missing")
			}

			var gotParentID string
			if err := json.Unmarshal(rawParentID, &gotParentID); err != nil {
				t.Fatalf("unmarshal parent_id: %v", err)
			}
			if gotParentID != tt.wantParentID {
				t.Errorf("parent_id = %q, want %q", gotParentID, tt.wantParentID)
			}
		})
	}
}
