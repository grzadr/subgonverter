package main

import (
	"testing"

	"github.com/grzadr/subgonverter/subtitle"
)

func TestParseArguments(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantConfig MainConfig
	}{
		{
			name: "no arguments - defaults",
			args: []string{},
			wantConfig: MainConfig{
				InputPath:    "",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "-",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "input file only",
			args: []string{"input.txt"},
			wantConfig: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "-",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "output flag with -o",
			args: []string{"-o", "output.srt"},
			wantConfig: MainConfig{
				InputPath:    "",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "output.srt",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "output flag with --output",
			args: []string{"--output", "output.srt"},
			wantConfig: MainConfig{
				InputPath:    "",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "output.srt",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "input and output with -o",
			args: []string{"-o", "output.srt", "input.txt"},
			wantConfig: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "output.srt",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "input and output with --output",
			args: []string{"--output", "output.srt", "input.txt"},
			wantConfig: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "output.srt",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "flags after positional args are treated as args",
			args: []string{"input.txt", "-o", "output.srt"},
			wantConfig: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "-",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "stdout explicitly with -o",
			args: []string{"-o", "-", "input.txt"},
			wantConfig: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "-",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "stdout explicitly with --output",
			args: []string{"--output", "-", "input.txt"},
			wantConfig: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "-",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "--output takes precedence over -o",
			args: []string{"-o", "file1.srt", "--output", "file2.srt"},
			wantConfig: MainConfig{
				InputPath:    "",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "file2.srt",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "multiple positional args - only first is used",
			args: []string{"input1.txt", "input2.txt"},
			wantConfig: MainConfig{
				InputPath:    "input1.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "-",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "file path with spaces",
			args: []string{"-o", "output file.srt", "input file.txt"},
			wantConfig: MainConfig{
				InputPath:    "input file.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "output file.srt",
				OutputFormat: subtitle.SrtFormat,
			},
		},
		{
			name: "file path with special characters",
			args: []string{"-o", "out-put_2024.srt", "in-put_2024.txt"},
			wantConfig: MainConfig{
				InputPath:    "in-put_2024.txt",
				InputFormat:  subtitle.TxtFormat,
				OutputPath:   "out-put_2024.srt",
				OutputFormat: subtitle.SrtFormat,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArguments(tt.args)

			if err != nil {
				t.Fatalf("ParseArguments() unexpected error: %v", err)
			}

			if got.InputPath != tt.wantConfig.InputPath {
				t.Errorf("InputPath = %q, want %q", got.InputPath, tt.wantConfig.InputPath)
			}
			if got.InputFormat != tt.wantConfig.InputFormat {
				t.Errorf("InputFormat = %v, want %v", got.InputFormat, tt.wantConfig.InputFormat)
			}
			if got.OutputPath != tt.wantConfig.OutputPath {
				t.Errorf("OutputPath = %q, want %q", got.OutputPath, tt.wantConfig.OutputPath)
			}
			if got.OutputFormat != tt.wantConfig.OutputFormat {
				t.Errorf("OutputFormat = %v, want %v", got.OutputFormat, tt.wantConfig.OutputFormat)
			}
		})
	}
}

func TestParseArgumentsErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "unknown flag",
			args: []string{"-x", "value"},
		},
		{
			name: "flag without value",
			args: []string{"-o"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArguments(tt.args)

			if err == nil {
				t.Error("ParseArguments() expected error but got nil")
			}
		})
	}
}
