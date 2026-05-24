package resolver

import (
	"testing"

	"github.com/SheltonZhu/115driver/pkg/driver"
)

func TestResolvePath_FallsBackToFileWhenDirLookupReturnsRootID(t *testing.T) {
	client := fakeResolverClient{
		dirIDs: map[string]string{
			"q9tVD1jYR8e626EteJ0qDQ.mp4": RootID,
		},
		pagesByDir: map[string][][]driver.File{
			RootID: {
				{
					{
						FileID:      "123456",
						Name:        "q9tVD1jYR8e626EteJ0qDQ.mp4",
						IsDirectory: false,
					},
				},
			},
		},
	}

	fileID, isDir, err := ResolvePath(&client, "q9tVD1jYR8e626EteJ0qDQ.mp4")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}
	if isDir {
		t.Fatalf("ResolvePath should treat file path as file")
	}
	if fileID != "123456" {
		t.Fatalf("unexpected file ID: %s", fileID)
	}
}

func TestResolvePath_RootStillResolvesToDirectory(t *testing.T) {
	fileID, isDir, err := ResolvePath(&fakeResolverClient{}, "/")
	if err != nil {
		t.Fatalf("ResolvePath returned error: %v", err)
	}
	if !isDir {
		t.Fatalf("root path should resolve as directory")
	}
	if fileID != RootID {
		t.Fatalf("unexpected root ID: %s", fileID)
	}
}

func TestResolveFileSearchesDirectoryByPages(t *testing.T) {
	client := fakeResolverClient{
		dirIDs: map[string]string{"movies": "dir1"},
		pagesByDir: map[string][][]driver.File{
			"dir1": {
				repeatFiles(fileResolvePageLimit, "filler"),
				{{FileID: "2", Name: "target.mp4"}},
			},
		},
	}

	fileID, err := ResolveFile(&client, "/movies/target.mp4")
	if err != nil {
		t.Fatalf("ResolveFile returned error: %v", err)
	}
	if fileID != "2" {
		t.Fatalf("unexpected file ID: %s", fileID)
	}
	if client.listAllCalls != 0 {
		t.Fatalf("expected ResolveFile not to call full List, got %d calls", client.listAllCalls)
	}
	if client.listPageCalls != 2 {
		t.Fatalf("expected ResolveFile to scan 2 pages, got %d", client.listPageCalls)
	}
}

type fakeResolverClient struct {
	dirIDs        map[string]string
	filesByDir    map[string][]driver.File
	pagesByDir    map[string][][]driver.File
	listAllCalls  int
	listPageCalls int
}

func (f fakeResolverClient) DirName2CID(dir string) (*driver.APIGetDirIDResp, error) {
	id := f.dirIDs[dir]
	return &driver.APIGetDirIDResp{
		CategoryID: driver.IntString(id),
	}, nil
}

func (f *fakeResolverClient) List(dirID string, _ ...driver.ListOption) (*[]driver.File, error) {
	f.listAllCalls++
	files := f.filesByDir[dirID]
	return &files, nil
}

func (f *fakeResolverClient) ListPage(dirID string, offset, limit int64, _ ...driver.ListOption) (*[]driver.File, error) {
	f.listPageCalls++
	page := int(offset / limit)
	pages := f.pagesByDir[dirID]
	if page >= len(pages) {
		files := []driver.File{}
		return &files, nil
	}
	files := pages[page]
	return &files, nil
}

func repeatFiles(count int64, prefix string) []driver.File {
	files := make([]driver.File, 0, count)
	for i := int64(0); i < count; i++ {
		files = append(files, driver.File{FileID: prefix, Name: prefix})
	}
	return files
}
