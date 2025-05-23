/*
Copyright 2025 Lucas Chagas <lucas.w.chagas@gmail.com>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/DiskMethod/cs2-voice-tools/internal/extract"
	"github.com/spf13/cobra"
)

var (
	// playerFilter is a comma-separated list of SteamID64s to filter by
	playerFilter string

	// formatOption specifies the output format for audio files
	formatOption string

	// steamID64Regex is the regular expression for validating SteamID64 format
	// SteamID64 should be a 17-digit number starting with 7656
	steamID64Regex = regexp.MustCompile(`^7656\d{13}$`)
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [flags] <demo-file>",
	Short: "Extract voice data from a CS2 demo",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		demoPath := args[0]

		// Parse player filter if provided
		var playerIDs []string
		var invalidIDs []string

		if playerFilter != "" {
			// Split the comma-separated list and trim whitespace
			for _, id := range strings.Split(playerFilter, ",") {
				// Trim whitespace and ensure non-empty
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}

				// Validate SteamID64 format
				if !steamID64Regex.MatchString(id) {
					slog.Warn("Invalid SteamID64 format, skipping", "id", id)
					invalidIDs = append(invalidIDs, id)
					continue
				}

				playerIDs = append(playerIDs, id)
			}

			// Warn if no valid IDs were provided
			if len(playerIDs) == 0 && len(invalidIDs) > 0 {
				return fmt.Errorf("no valid SteamID64s provided, received: %s", strings.Join(invalidIDs, ", "))
			}
		}

		// Validate format option
		format := strings.ToLower(formatOption)
		isFormatValid := false

		if format == "" {
			// Default to WAV if no format specified
			format = "wav"
			isFormatValid = true
		} else {
			// Check if the format is supported
			for _, supportedFormat := range extract.GetSupportedFormats() {
				if format == supportedFormat {
					isFormatValid = true
					break
				}
			}
		}

		if !isFormatValid {
			return fmt.Errorf("unsupported format: %s (supported formats: %s)",
				format, strings.Join(extract.GetSupportedFormats(), ", "))
		}

		// Create extract options from command-line arguments
		options := extract.ExtractOptions{
			DemoPath:       demoPath,
			OutputDir:      Opts.AbsOutputDir,
			ForceOverwrite: Opts.ForceOverwrite,
			PlayerIDs:      playerIDs,
			Format:         format,
		}

		// Extract voice data with the configured options
		if err := extract.ExtractVoiceData(options); err != nil {
			return err
		}

		msg := fmt.Sprintf("Voice data extraction complete. Files saved to: %s", Opts.AbsOutputDir)
		if len(playerIDs) > 0 {
			msg += fmt.Sprintf(" (filtered to %d players)", len(playerIDs))
		}
		if format != "wav" {
			msg += fmt.Sprintf(" (format: %s)", format)
		}
		fmt.Println(msg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)

	// Add command-specific flags
	extractCmd.Flags().StringVarP(&playerFilter, "players", "p", "", "filter to specific players by steamID64 (comma-separated list)")
	extractCmd.Flags().StringVarP(&formatOption, "format", "t", "wav",
		fmt.Sprintf("output audio format (%s)", strings.Join(extract.GetSupportedFormats(), ", ")))
}
