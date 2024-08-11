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

func (f *File) From(fileInfo *FileInfo) *File {
	return f.from(fileInfo)
}

func (f *File) from(fileInfo *FileInfo) *File {
	if fileInfo.FileID != "" {
		f.FileID = fileInfo.FileID
		f.ParentID = string(fileInfo.CategoryID)
		f.IsDirectory = false
		loc, err := time.LoadLocation("Asia/Shanghai") // updatetime is a string without timezone
		if err != nil {
			// if missing Asia/Shanghai use CST（UTC+8） 
			 loc = time.FixedZone("UTC+8", 8*3600)
		}
		localTime, err := time.ParseInLocation("2006-01-02 15:04", fileInfo.UpdateTime, loc)
		if err == nil {
			f.UpdateTime = time.Unix(localTime.Unix(), 0)
		}
	} else {
		f.FileID = string(fileInfo.CategoryID)
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

func (f File) GetPath() string {
	return ""
}

func (f File) GetSize() int64 {
	return f.Size
}

func (f File) GetName() string {
	return f.Name
}

func (f File) ModTime() time.Time {
	return f.UpdateTime
}

func (f File) IsDir() bool {
	return f.IsDirectory
}

func (f File) GetID() string {
	return f.FileID
}
