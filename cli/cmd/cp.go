package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cpCmd = &cobra.Command{
	Use:   "cp <source_path>... <destination_dir>",
	Short: "Copy files into a destination directory",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires one or more <source_path> arguments and a <destination_dir>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return moveOrCopy(args[:len(args)-1], args[len(args)-1], client.Copy)
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
