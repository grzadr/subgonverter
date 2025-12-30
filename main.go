package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"os/signal"
)

const bufferSize = 256 * 1024

var ErrNotImplemented = errors.New("not implemented")

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

type MainConfig struct {
	InputPath    string
	InputFormat  FileFormat
	OutputFormat FileFormat
}

type Subtitle struct{}

func ParseArguments(args []string) (parsed MainConfig, err error) {
	return parsed, err
}

func InitReader(path string) (reader io.Reader, err error) {
	return reader, err
}

func TransformInput(
	ctx context.Context,
	reader io.Reader,
) iter.Seq2[Subtitle, error] {
	return func(yield func(Subtitle, error) bool) {
		return
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config, err := ParseArguments(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to parse arguments: %s", err)
	}

	reader, err := InitReader(config.InputPath)
	if err != nil {
		log.Fatalf("failed to initialize input read: %s", err)
	}
}
