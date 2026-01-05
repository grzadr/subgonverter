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
	"time"
)

const bufferSize = 256 * 1024

var errNotImplemented = errors.New("not implemented")

type FileFormat uint8

const (
	UnknownFormat FileFormat = iota
	TxtFormat
	SrtFormat
)

type MainConfig struct {
	InputPath    string
	InputFormat  FileFormat
	OutputPath   string
	OutputFormat FileFormat
}

type Subtitle struct {
	Start time.Duration
	End   time.Duration
	Text  string
}

func NewSubtitleFromTxt(line string) (sub Subtitle, err error) {
	return sub, errNotImplemented
}

func ParseArguments(args []string) (parsed MainConfig, err error) {
	return parsed, errNotImplemented
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

	bw := bufio.NewWriterSize(file, bufferSize)

	cleanup := func() error {
		if err := bw.Flush(); err != nil {
			file.Close()
			return err
		}
		return file.Close()
	}

	return bw, cleanup, nil
}

func NewScannerPull(reader io.Reader) (
	next func() (string, error, bool),
	stop func(),
) {
	scanner := bufio.NewScanner(reader)

	buf := make([]byte, 0, bufferSize)
	scanner.Buffer(buf, bufferSize)

	return iter.Pull2(func(yield func(string, error) bool) {
		for scanner.Scan() {
			if !yield(scanner.Text(), nil) {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			yield("", err)
		}
	})
}

func NewTxtSubtitlesIter(
	next func() (string, error, bool),
	stop func(),
) iter.Seq2[Subtitle, error] {
	return func(yield func(Subtitle, error) bool) {
		defer stop()

		for {
			line, err, ok := next()
			if !ok {
				return
			}
			if err != nil {
				yield(
					Subtitle{},
					fmt.Errorf("error reading txt subtitle: %w", err),
				)
				return
			}

			sub, err := NewSubtitleFromTxt(line)
			if err != nil {
				yield(
					Subtitle{},
					fmt.Errorf("error parsing txt subtitle: %w", err),
				)
				return
			}

			if !yield(sub, err) {
				return
			}
		}
	}
}

func NewSubtitlesIter(
	reader io.Reader,
	format FileFormat,
) iter.Seq2[Subtitle, error] {
	next, stop := NewScannerPull(reader)

	switch format {
	case TxtFormat:
		return NewTxtSubtitlesIter(next, stop)

	default:
		return func(yield func(Subtitle, error) bool) {
			defer stop()

			yield(Subtitle{}, errNotImplemented)
		}
	}
}

func NewSubtitleWriter(
	writer io.Writer,
	format FileFormat,
) (func(sub Subtitle) error, error) {
	switch format {
	case SrtFormat:

		return
	case TxtFormat:
		return errNotImplemented
	default:
		return errors.New("unknown output file format")
	}
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

	for sub, err := range NewSubtitlesIter(reader, config.InputFormat) {

		if err != nil {
			return fmt.Errorf("failed to parse subtitle: %s", err)
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if err := WriteSubtitle(writer, sub, config.OutputFormat); err != nil {
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
