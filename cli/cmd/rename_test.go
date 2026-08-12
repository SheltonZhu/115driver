package cmd

import "testing"

func TestRenameArgs(t *testing.T) {
	for _, args := range [][]string{
		{"/a", "new-a"},
		{"/a", "new-a", "/b", "new-b"},
	} {
		if err := renameCmd.Args(renameCmd, args); err != nil {
			t.Fatalf("valid args %v rejected: %v", args, err)
		}
	}

	for _, args := range [][]string{
		nil,
		{"/a"},
		{"/a", "new-a", "/b"},
	} {
		if err := renameCmd.Args(renameCmd, args); err == nil {
			t.Fatalf("invalid args %v accepted", args)
		}
	}
}
