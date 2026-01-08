package subtitle

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
	"strconv"
	"strings"
	"time"
)

const (
	ntscRateNum = 24000
	ntscRateDen = 1001
	ntscRateDiv = 1000
)

type Subtitle struct {
	Start time.Duration
	End   time.Duration
	Text  string
}

func newSubtitleFromTxt(line string) (sub Subtitle, err error) {
	// Parse format: {123}{164}text|text
	// Assume input is correctly formatted

	// Find positions of braces (adding +1 to skip past the braces)
	startFrom := strings.Index(line, "{") + 1
	startTill := strings.Index(line, "}")
	endFrom := strings.Index(line[startTill:], "{") + startTill + 1
	endTill := strings.Index(line[endFrom-1:], "}") + endFrom - 1

	// Parse frame numbers
	startFrame, err := strconv.ParseInt(line[startFrom:startTill], 10, 64)
	if err != nil {
		return sub, fmt.Errorf("failed to parse start frame: %w", err)
	}

	endFrame, err := strconv.ParseInt(line[endFrom:endTill], 10, 64)
	if err != nil {
		return sub, fmt.Errorf("failed to parse end frame: %w", err)
	}

	// Extract text and convert | to newlines
	text := strings.ReplaceAll(line[endTill+1:], "|", "\n")

	// Convert frames to duration using NTSC rate
	// frame * (ntscRateDen * ntscRateDiv) / ntscRateNum = milliseconds
	div := int64(ntscRateNum)
	mul := int64(ntscRateDen * ntscRateDiv)

	sub.Start = time.Duration(startFrame*mul/div) * time.Millisecond
	sub.End = time.Duration(endFrame*mul/div) * time.Millisecond
	sub.Text = text

	return sub, nil
}

func writeTxtDuration(w io.Writer, d time.Duration) error {
	div := int64(ntscRateDen * ntscRateDiv)
	frame := (d.Milliseconds()*ntscRateNum + div/2) / div
	_, err := fmt.Fprintf(
		w,
		"{%d}",
		frame,
	)

	return err
}

func writeSrtDuration(w io.Writer, d time.Duration) error {
	hours := d / time.Hour
	minutes := (d % time.Hour) / time.Minute
	seconds := (d % time.Minute) / time.Second
	millis := (d % time.Second) / time.Millisecond
	_, err := fmt.Fprintf(
		w,
		"%02d:%02d:%02d,%03d",
		hours,
		minutes,
		seconds,
		millis,
	)

	return err
}

func writeSrtSubtitle(w io.Writer, sub Subtitle, n int) error {
	var err error

	if _, err = fmt.Fprintln(w, n); err != nil {
		return err
	}

	if err = writeSrtDuration(w, sub.Start); err != nil {
		return err
	}

	if _, err = fmt.Fprint(w, " --> "); err != nil {
		return err
	}

	if err = writeSrtDuration(w, sub.Start); err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, "\n", sub.Text)
	return err
}

var ErrNotImplemented = errors.New("not implemented")

const readBufferSize = 256 * 1024

type FileFormat uint8

const (
	UnknownFormat FileFormat = iota
	TxtFormat
	SrtFormat
)

func NewSubtitlePrinter(
	writer io.Writer,
	format FileFormat,
) func(sub Subtitle) error {
	switch format {
	case SrtFormat:
		n := 0
		return func(sub Subtitle) error {
			n++
			return writeSrtSubtitle(writer, sub, n)
		}
	case TxtFormat:
		return nil
	default:
		return nil
	}
}

func newScannerPull(reader io.Reader) (
	next func() (string, error, bool),
	stop func(),
) {
	scanner := bufio.NewScanner(reader)

	buf := make([]byte, 0, readBufferSize)
	scanner.Buffer(buf, readBufferSize)

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

func newTxtSubtitlesIter(
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

			sub, err := newSubtitleFromTxt(line)
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
	next, stop := newScannerPull(reader)

	switch format {
	case TxtFormat:
		return newTxtSubtitlesIter(next, stop)

	default:
		return func(yield func(Subtitle, error) bool) {
			defer stop()

			yield(Subtitle{}, ErrNotImplemented)
		}
	}
}
