package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
)

const bufferSize = 256 * 1024

func processFileLines[R any](
	filename string,
	process func(string) (R, error),
) (R, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	buf := make([]byte, 0, bufferSize)
	scanner.Buffer(buf, bufferSize)

	for scanner.Scan() {
		if err := process(scanner.Text()); err != nil {
			return fmt.Errorf("failed to process line: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

type FileFormat uint8

const (
	UnknownFormat FileFormat = iota
	TxtFormat
	SrtFormat
)

type Arguments struct {
	InputPath    string
	InputFormat  FileFormat
	OutputFormat FileFormat
}

func ParseArguments() (args Arguments, err error) {
	return args, err
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
}
