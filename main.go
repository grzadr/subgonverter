package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/grzadr/subgonverter/subtitle"
)

const (
	writeBufferSize = 256 * 1024
)

type MainConfig struct {
	InputPath    string
	InputFormat  subtitle.FileFormat
	OutputPath   string
	OutputFormat subtitle.FileFormat
}

func ParseArguments(args []string) (parsed MainConfig, err error) {
	return parsed, subtitle.ErrNotImplemented
}

func InitReader(path string) (io.Reader, func() error, error) {
	if path == "" || path == "-" {
		return os.Stdin, func() error { return nil }, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	return file, file.Close, nil
}

func InitWriter(path string) (io.Writer, func() error, error) {
	if path == "" || path == "-" {
		bw := bufio.NewWriter(os.Stdout)
		return bw, bw.Flush, nil
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}

	bw := bufio.NewWriterSize(file, writeBufferSize)

	cleanup := func() error {
		if err := bw.Flush(); err != nil {
			file.Close()
			return err
		}
		return file.Close()
	}

	return bw, cleanup, nil
}

func process(
	ctx context.Context,
	config MainConfig,
) error {
	reader, rcloser, err := InitReader(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to initialize input reader: %w", err)
	}
	defer rcloser()

	writer, wcloser, err := InitWriter(config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to initialize output writer: %w", err)
	}
	defer wcloser()

	print := subtitle.NewSubtitlePrinter(writer, config.OutputFormat)

	for sub, err := range subtitle.NewSubtitlesIter(reader, config.InputFormat) {

		if err != nil {
			return fmt.Errorf("failed to parse subtitle: %s", err)
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if err := print(sub); err != nil {
			return fmt.Errorf("failed to write subtitle: %w", err)
		}
	}

	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config, err := ParseArguments(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to parse arguments: %s", err)
	}

	if err := process(ctx, config); err != nil {
		log.Fatalf("processing failed: %v", err)
	}
}
