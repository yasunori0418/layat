package main

import "testing"

// TestApplyBackupFlagParsing: bare --backup uses the default suffix, --backup=<suffix> takes the
// given one, and a bare next token stays a positional argument.
func TestApplyBackupFlagParsing(t *testing.T) {
	t.Run("absent: Changed is false", func(t *testing.T) {
		cmd := newApplyCmd()
		if err := cmd.ParseFlags([]string{"name"}); err != nil {
			t.Fatalf("ParseFlags: %v", err)
		}
		if cmd.Flags().Changed("backup") {
			t.Error("Changed(\"backup\") = true, want false when --backup is absent")
		}
	})

	t.Run("bare --backup uses the default suffix", func(t *testing.T) {
		cmd := newApplyCmd()
		if err := cmd.ParseFlags([]string{"--backup", "name"}); err != nil {
			t.Fatalf("ParseFlags: %v", err)
		}
		if !cmd.Flags().Changed("backup") {
			t.Error("Changed(\"backup\") = false, want true")
		}
		got, err := cmd.Flags().GetString("backup")
		if err != nil {
			t.Fatalf("GetString(backup): %v", err)
		}
		if got != "layat-backup" {
			t.Errorf("bare --backup value = %q, want %q (NoOptDefVal)", got, "layat-backup")
		}
		// The bare form must not take the next token as its value.
		if args := cmd.Flags().Args(); len(args) != 1 || args[0] != "name" {
			t.Errorf("positional args after bare --backup = %v, want [name] (next token must not be consumed as the flag value)", args)
		}
	})

	t.Run("--backup=<suffix> takes the custom suffix", func(t *testing.T) {
		cmd := newApplyCmd()
		if err := cmd.ParseFlags([]string{"--backup=bak", "name"}); err != nil {
			t.Fatalf("ParseFlags: %v", err)
		}
		if !cmd.Flags().Changed("backup") {
			t.Error("Changed(\"backup\") = false, want true")
		}
		got, err := cmd.Flags().GetString("backup")
		if err != nil {
			t.Fatalf("GetString(backup): %v", err)
		}
		if got != "bak" {
			t.Errorf("--backup=bak value = %q, want %q", got, "bak")
		}
	})
}
