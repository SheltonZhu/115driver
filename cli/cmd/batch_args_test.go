package cmd

import "testing"

func TestBatchTransferArgs(t *testing.T) {
	for _, command := range []*struct {
		name string
		args func(*testing.T, []string) error
	}{
		{name: "mv", args: func(t *testing.T, args []string) error { return mvCmd.Args(mvCmd, args) }},
		{name: "cp", args: func(t *testing.T, args []string) error { return cpCmd.Args(cpCmd, args) }},
	} {
		if err := command.args(t, []string{"/a", "/dest"}); err != nil {
			t.Fatalf("%s rejected one source: %v", command.name, err)
		}
		if err := command.args(t, []string{"/a", "/b", "/dest"}); err != nil {
			t.Fatalf("%s rejected multiple sources: %v", command.name, err)
		}
		if err := command.args(t, []string{"/dest"}); err == nil {
			t.Fatalf("%s accepted a destination without a source", command.name)
		}
	}
}

func TestBatchDeleteArgs(t *testing.T) {
	if err := rmCmd.Args(rmCmd, []string{"/a", "/b"}); err != nil {
		t.Fatalf("rm rejected multiple paths: %v", err)
	}
	if err := rmCmd.Args(rmCmd, nil); err == nil {
		t.Fatal("rm accepted no paths")
	}
}

func TestBatchOfflineArgs(t *testing.T) {
	for _, command := range []*struct {
		name string
		args func([]string) error
	}{
		{name: "add", args: func(args []string) error { return offlineAddCmd.Args(offlineAddCmd, args) }},
		{name: "rm", args: func(args []string) error { return offlineRmCmd.Args(offlineRmCmd, args) }},
	} {
		if err := command.args([]string{"one", "two"}); err != nil {
			t.Fatalf("offline %s rejected multiple arguments: %v", command.name, err)
		}
		if err := command.args(nil); err == nil {
			t.Fatalf("offline %s accepted no arguments", command.name)
		}
	}
}

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
