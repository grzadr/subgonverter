package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestInitReader(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "stdin with empty string",
			path:    "",
			wantErr: false,
		},
		{
			name:    "stdin with dash",
			path:    "-",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, cleanup, err := InitReader(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("InitReader() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("InitReader() unexpected error: %v", err)
			}

			if reader == nil {
				t.Error("InitReader() returned nil reader")
			}

			if cleanup == nil {
				t.Error("InitReader() returned nil cleanup function")
			}

			// Verify cleanup doesn't error
			if err := cleanup(); err != nil {
				t.Errorf("cleanup() unexpected error: %v", err)
			}

			// For stdin tests, verify we got os.Stdin
			if tt.path == "" || tt.path == "-" {
				if reader != os.Stdin {
					t.Error("InitReader() expected os.Stdin for stdin path")
				}
			}
		})
	}
}

func TestInitReader_WithFile(t *testing.T) {
	// Create a temporary file with test content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_input.txt")
	testContent := "{100}{200}Test subtitle\n"

	if err := os.WriteFile(tmpFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	reader, cleanup, err := InitReader(tmpFile)
	if err != nil {
		t.Fatalf("InitReader() unexpected error: %v", err)
	}

	if reader == nil {
		t.Fatal("InitReader() returned nil reader")
	}

	if cleanup == nil {
		t.Fatal("InitReader() returned nil cleanup function")
	}

	// Read and verify content
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read from reader: %v", err)
	}

	if string(data) != testContent {
		t.Errorf("expected content %q, got %q", testContent, string(data))
	}

	// Cleanup
	if err := cleanup(); err != nil {
		t.Errorf("cleanup() unexpected error: %v", err)
	}
}

func TestInitReader_NonExistentFile(t *testing.T) {
	reader, cleanup, err := InitReader("/nonexistent/file/path.txt")

	if err == nil {
		t.Error("InitReader() expected error for non-existent file, got nil")
	}

	if reader != nil {
		t.Error("InitReader() expected nil reader for error case")
	}

	if cleanup != nil {
		t.Error("InitReader() expected nil cleanup for error case")
	}
}

func TestInitWriter(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "stdout with empty string",
			path:    "",
			wantErr: false,
		},
		{
			name:    "stdout with dash",
			path:    "-",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, cleanup, err := InitWriter(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("InitWriter() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("InitWriter() unexpected error: %v", err)
			}

			if writer == nil {
				t.Error("InitWriter() returned nil writer")
			}

			if cleanup == nil {
				t.Error("InitWriter() returned nil cleanup function")
			}

			// Verify cleanup doesn't error (this will flush to stdout)
			if err := cleanup(); err != nil {
				t.Errorf("cleanup() unexpected error: %v", err)
			}
		})
	}
}

func TestInitWriter_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_output.txt")
	testContent := "Test output content"

	writer, cleanup, err := InitWriter(tmpFile)
	if err != nil {
		t.Fatalf("InitWriter() unexpected error: %v", err)
	}

	if writer == nil {
		t.Fatal("InitWriter() returned nil writer")
	}

	if cleanup == nil {
		t.Fatal("InitWriter() returned nil cleanup function")
	}

	// Write test content
	n, err := writer.Write([]byte(testContent))
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	if n != len(testContent) {
		t.Errorf("wrote %d bytes, expected %d", n, len(testContent))
	}

	// Cleanup (this should flush and close)
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup() unexpected error: %v", err)
	}

	// Verify file content
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(data) != testContent {
		t.Errorf("expected content %q, got %q", testContent, string(data))
	}
}

func TestInitWriter_InvalidPath(t *testing.T) {
	writer, cleanup, err := InitWriter("/nonexistent/directory/output.txt")

	if err == nil {
		t.Error("InitWriter() expected error for invalid path, got nil")
	}

	if writer != nil {
		t.Error("InitWriter() expected nil writer for error case")
	}

	if cleanup != nil {
		t.Error("InitWriter() expected nil cleanup for error case")
	}
}

func TestInitWriter_MultipleWrites(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_multiple_writes.txt")

	writer, cleanup, err := InitWriter(tmpFile)
	if err != nil {
		t.Fatalf("InitWriter() unexpected error: %v", err)
	}

	// Write multiple times
	writes := []string{"First ", "second ", "third"}
	expectedContent := strings.Join(writes, "")

	for _, s := range writes {
		if _, err := writer.Write([]byte(s)); err != nil {
			t.Fatalf("failed to write: %v", err)
		}
	}

	// Cleanup
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup() unexpected error: %v", err)
	}

	// Verify all content was written
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(data) != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, string(data))
	}
}

func TestInitWriter_FlushAndClose(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_flush_close.txt")

	writer, cleanup, err := InitWriter(tmpFile)
	if err != nil {
		t.Fatalf("InitWriter() unexpected error: %v", err)
	}

	// Write buffered content without flushing first
	testData := strings.Repeat("x", writeBufferSize/2) // Less than buffer size
	if _, err := writer.Write([]byte(testData)); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Cleanup should flush and close
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup() unexpected error: %v", err)
	}

	// Verify content was flushed
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(data) != testData {
		t.Errorf("expected data length %d, got %d", len(testData), len(data))
	}

	// Calling cleanup again should be safe (though may return error since file is closed)
	_ = cleanup()
}
