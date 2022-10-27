package driver

import (
	"strconv"
	"time"
)

type File struct {
	// Marks is the file a directory.
	IsDirectory bool
	// Unique identifier of the file on the cloud storage.
	FileID string
	// FileID of the parent directory.
	ParentID string

	// Base name of the file.
	Name string
	// Size in bytes of the file.
	Size int64
	// IDentifier used for downloading or playing the file.
	PickCode string
	// SHA1 hash of file content, in HEX format.
	Sha1 string

	// Is file stared
	Star bool
	// File labels
	Labels []*Label

	// Create time of the file.
	CreateTime time.Time
	// Update time of the file.
	UpdateTime time.Time
}

func (f *File) from(fileInfo *FileInfo) *File {
	if fileInfo.FileID != "" {
		f.FileID = fileInfo.FileID
		f.ParentID = fileInfo.CategoryID
		f.IsDirectory = false
		t, err := time.Parse("2006-01-02 15:04", fileInfo.UpdateTime)
		if err == nil {
			f.UpdateTime = time.Unix(t.Unix(), 0)
		}
	} else {
		f.FileID = fileInfo.CategoryID
		f.ParentID = fileInfo.ParentID
		f.IsDirectory = true
		t, err := strconv.ParseInt(fileInfo.UpdateTime, 10, 64)
		if err == nil {
			f.UpdateTime = time.Unix(t, 0)
		}
	}
	f.Name = fileInfo.Name
	f.Size = int64(fileInfo.Size)
	f.PickCode = fileInfo.PickCode
	f.Sha1 = fileInfo.Sha1

	f.Star = fileInfo.IsStar != 0
	f.Labels = make([]*Label, len(fileInfo.Labels))
	for i, l := range fileInfo.Labels {
		f.Labels[i] = &Label{
			ID:    l.ID,
			Name:  l.Name,
			Color: LabelColor(LabelColorMap[l.Color]),
		}
	}

	f.CreateTime = time.Unix(int64(fileInfo.CreateTime), 0)

	return f
}
