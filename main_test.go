package main

import (
	"testing"
)

func TestParseArguments(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		want        MainConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "empty arguments",
			args: []string{},
			want: MainConfig{
				InputPath:    "",
				InputFormat:  UnknownFormat,
				OutputFormat: UnknownFormat,
			},
			wantErr:     true,
			errContains: "",
		},
		{
			name: "single input file argument",
			args: []string{"input.txt"},
			want: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  UnknownFormat,
				OutputFormat: UnknownFormat,
			},
			wantErr: false,
		},
		{
			name: "input file with txt format",
			args: []string{"--input-format", "txt", "input.txt"},
			want: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  TxtFormat,
				OutputFormat: UnknownFormat,
			},
			wantErr: false,
		},
		{
			name: "input file with srt format",
			args: []string{"--input-format", "srt", "input.srt"},
			want: MainConfig{
				InputPath:    "input.srt",
				InputFormat:  SrtFormat,
				OutputFormat: UnknownFormat,
			},
			wantErr: false,
		},
		{
			name: "input and output formats",
			args: []string{"--input-format", "txt", "--output-format", "srt", "input.txt"},
			want: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  TxtFormat,
				OutputFormat: SrtFormat,
			},
			wantErr: false,
		},
		{
			name: "output format only",
			args: []string{"--output-format", "srt", "input.txt"},
			want: MainConfig{
				InputPath:    "input.txt",
				InputFormat:  UnknownFormat,
				OutputFormat: SrtFormat,
			},
			wantErr: false,
		},
		{
			name:        "invalid input format",
			args:        []string{"--input-format", "invalid", "input.txt"},
			want:        MainConfig{},
			wantErr:     true,
			errContains: "format",
		},
		{
			name:        "invalid output format",
			args:        []string{"--output-format", "invalid", "input.txt"},
			want:        MainConfig{},
			wantErr:     true,
			errContains: "format",
		},
		{
			name:        "missing input file",
			args:        []string{"--input-format", "txt"},
			want:        MainConfig{},
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "unknown flag",
			args:        []string{"--unknown-flag", "value", "input.txt"},
			want:        MainConfig{},
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArguments(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseArguments() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("ParseArguments() error = %v, want error containing %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("ParseArguments() unexpected error = %v", err)
				return
			}

			if got.InputPath != tt.want.InputPath {
				t.Errorf("ParseArguments().InputPath = %v, want %v", got.InputPath, tt.want.InputPath)
			}
			if got.InputFormat != tt.want.InputFormat {
				t.Errorf("ParseArguments().InputFormat = %v, want %v", got.InputFormat, tt.want.InputFormat)
			}
			if got.OutputFormat != tt.want.OutputFormat {
				t.Errorf("ParseArguments().OutputFormat = %v, want %v", got.OutputFormat, tt.want.OutputFormat)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
