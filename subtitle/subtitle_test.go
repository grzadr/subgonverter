package subtitle

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewSubtitlesIter_TxtFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []struct {
			startFrame int64
			endFrame   int64
			text       string
		}
	}{
		{
			name:  "single subtitle",
			input: "{123}{164}Hello World",
			want: []struct {
				startFrame int64
				endFrame   int64
				text       string
			}{
				{123, 164, "Hello World"},
			},
		},
		{
			name: "multiple subtitles",
			input: `{100}{150}First subtitle
{200}{250}Second subtitle
{300}{350}Third subtitle`,
			want: []struct {
				startFrame int64
				endFrame   int64
				text       string
			}{
				{100, 150, "First subtitle"},
				{200, 250, "Second subtitle"},
				{300, 350, "Third subtitle"},
			},
		},
		{
			name:  "multiline text with pipe separator",
			input: "{500}{600}Line one|Line two|Line three",
			want: []struct {
				startFrame int64
				endFrame   int64
				text       string
			}{
				{500, 600, "Line one\nLine two\nLine three"},
			},
		},
		{
			name:  "zero frames",
			input: "{0}{0}Instant subtitle",
			want: []struct {
				startFrame int64
				endFrame   int64
				text       string
			}{
				{0, 0, "Instant subtitle"},
			},
		},
		{
			name:  "text containing pipe within multiline",
			input: "{100}{200}First line|Second | with pipe|Third",
			want: []struct {
				startFrame int64
				endFrame   int64
				text       string
			}{
				{100, 200, "First line\nSecond \n with pipe\nThird"},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  []struct {
				startFrame int64
				endFrame   int64
				text       string
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			iter := NewSubtitlesIter(reader, TxtFormat)

			count := 0
			for sub, err := range iter {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if count >= len(tt.want) {
					t.Fatalf("got more subtitles than expected")
				}

				exp := tt.want[count]

				if sub.Text != exp.text {
					t.Errorf("subtitle %d: expected text %q, got %q", count, exp.text, sub.Text)
				}

				expectedStart := time.Duration(exp.startFrame*1001*1000/24000) * time.Millisecond
				expectedEnd := time.Duration(exp.endFrame*1001*1000/24000) * time.Millisecond

				if sub.Start != expectedStart {
					t.Errorf("subtitle %d: expected start %v, got %v", count, expectedStart, sub.Start)
				}

				if sub.End != expectedEnd {
					t.Errorf("subtitle %d: expected end %v, got %v", count, expectedEnd, sub.End)
				}

				count++
			}

			if count != len(tt.want) {
				t.Errorf("expected %d subtitles, got %d", len(tt.want), count)
			}
		})
	}
}

func TestNewSubtitlesIter_UnknownFormat(t *testing.T) {
	input := "{100}{150}Some text"
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, UnknownFormat)

	count := 0
	for _, err := range iter {
		count++
		if err != ErrNotImplemented {
			t.Errorf("expected ErrNotImplemented, got %v", err)
		}
	}

	if count != 1 {
		t.Errorf("expected iterator to yield error once, got %d", count)
	}
}

func TestNewSubtitlesIter_SrtFormat(t *testing.T) {
	input := "1\n00:00:05,120 --> 00:00:06,840\nSome SRT subtitle"
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, SrtFormat)

	count := 0
	for _, err := range iter {
		count++
		// SRT format is not implemented, should return ErrNotImplemented
		if err != ErrNotImplemented {
			t.Errorf("expected ErrNotImplemented for SrtFormat, got %v", err)
		}
	}

	if count != 1 {
		t.Errorf("expected iterator to yield error once, got %d", count)
	}
}

func TestNewSubtitlesIter_TxtFormat_EarlyBreak(t *testing.T) {
	// Test that the iterator stops when we break early
	input := `{100}{150}First subtitle
{200}{250}Second subtitle
{300}{350}Third subtitle
{400}{450}Fourth subtitle`
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, TxtFormat)

	count := 0
	for sub, err := range iter {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		count++

		// Break after reading 2 subtitles
		if count == 2 {
			// Verify we got the second subtitle
			if sub.Text != "Second subtitle" {
				t.Errorf("expected 'Second subtitle', got '%s'", sub.Text)
			}
			break
		}
	}

	if count != 2 {
		t.Errorf("expected to process 2 subtitles before break, got %d", count)
	}
}

func TestNewSubtitleFromTxt_LargeFrameNumbers(t *testing.T) {
	// Test with large frame numbers to ensure proper calculation
	input := "{100000}{200000}Large frame numbers"
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, TxtFormat)

	for sub, err := range iter {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedStart := time.Duration(100000*1001*1000/24000) * time.Millisecond
		expectedEnd := time.Duration(200000*1001*1000/24000) * time.Millisecond

		if sub.Start != expectedStart {
			t.Errorf("expected start %v, got %v", expectedStart, sub.Start)
		}

		if sub.End != expectedEnd {
			t.Errorf("expected end %v, got %v", expectedEnd, sub.End)
		}

		if sub.Text != "Large frame numbers" {
			t.Errorf("expected text 'Large frame numbers', got '%s'", sub.Text)
		}
	}
}

func TestNewSubtitlesIter_TxtFormat_EmptyText(t *testing.T) {
	// Test subtitle with no text after frame numbers
	input := "{100}{200}"
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, TxtFormat)

	count := 0
	for sub, err := range iter {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		count++

		if sub.Text != "" {
			t.Errorf("expected empty text, got '%s'", sub.Text)
		}

		expectedStart := time.Duration(100*1001*1000/24000) * time.Millisecond
		expectedEnd := time.Duration(200*1001*1000/24000) * time.Millisecond

		if sub.Start != expectedStart {
			t.Errorf("expected start %v, got %v", expectedStart, sub.Start)
		}

		if sub.End != expectedEnd {
			t.Errorf("expected end %v, got %v", expectedEnd, sub.End)
		}
	}

	if count != 1 {
		t.Errorf("expected 1 subtitle, got %d", count)
	}
}

func TestNewSubtitlesIter_TxtFormat_OnlyPipes(t *testing.T) {
	// Test subtitle with only pipe characters (should become newlines)
	input := "{100}{200}|||"
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, TxtFormat)

	for sub, err := range iter {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedText := "\n\n\n"
		if sub.Text != expectedText {
			t.Errorf("expected text %q, got %q", expectedText, sub.Text)
		}
	}
}

func TestNewSubtitlesIter_TxtFormat_InvalidStartFrame(t *testing.T) {
	// Test with invalid start frame (non-numeric)
	input := "{abc}{200}Invalid start"
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, TxtFormat)

	count := 0
	for _, err := range iter {
		count++
		if err == nil {
			t.Fatal("expected error for invalid start frame, got nil")
		}

		if !strings.Contains(err.Error(), "failed to parse start frame") {
			t.Errorf("expected 'failed to parse start frame' error, got: %v", err)
		}
	}

	if count != 1 {
		t.Errorf("expected iterator to yield error once, got %d", count)
	}
}

func TestNewSubtitlesIter_TxtFormat_InvalidEndFrame(t *testing.T) {
	// Test with invalid end frame (non-numeric)
	input := "{100}{xyz}Invalid end"
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, TxtFormat)

	count := 0
	for _, err := range iter {
		count++
		if err == nil {
			t.Fatal("expected error for invalid end frame, got nil")
		}

		if !strings.Contains(err.Error(), "failed to parse end frame") {
			t.Errorf("expected 'failed to parse end frame' error, got: %v", err)
		}
	}

	if count != 1 {
		t.Errorf("expected iterator to yield error once, got %d", count)
	}
}

// errorReader is a custom reader that returns an error after reading some data
type errorReader struct {
	data    string
	pos     int
	readErr error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, r.readErr
	}

	n = copy(p, r.data[r.pos:])
	r.pos += n

	if r.pos >= len(r.data) {
		return n, r.readErr
	}

	return n, nil
}

func TestNewSubtitlesIter_TxtFormat_ReaderError(t *testing.T) {
	// Test with a reader that returns an error
	customErr := errors.New("reader error")
	reader := &errorReader{
		data:    "{100}{200}First subtitle\n",
		readErr: customErr,
	}

	iter := NewSubtitlesIter(reader, TxtFormat)

	count := 0
	gotError := false
	for _, err := range iter {
		count++
		if err != nil {
			gotError = true
			if !strings.Contains(err.Error(), "error reading txt subtitle") {
				t.Errorf("expected 'error reading txt subtitle' error, got: %v", err)
			}
		}
	}

	if !gotError {
		t.Error("expected to receive reader error")
	}
}

func TestNewSubtitlesIter_TxtFormat_ParseErrorAfterValidSubtitle(t *testing.T) {
	// Test with a valid subtitle followed by an invalid one
	input := `{100}{200}Valid subtitle
{invalid}{300}Bad start frame`
	reader := strings.NewReader(input)

	iter := NewSubtitlesIter(reader, TxtFormat)

	count := 0
	gotValid := false
	gotError := false

	for sub, err := range iter {
		count++

		if err != nil {
			gotError = true
			if !strings.Contains(err.Error(), "error parsing txt subtitle") {
				t.Errorf("expected 'error parsing txt subtitle' error, got: %v", err)
			}
			break
		}

		if sub.Text == "Valid subtitle" {
			gotValid = true
		}
	}

	if !gotValid {
		t.Error("expected to receive valid subtitle before error")
	}

	if !gotError {
		t.Error("expected to receive parse error for invalid subtitle")
	}

	if count != 2 {
		t.Errorf("expected 2 iterations (1 valid + 1 error), got %d", count)
	}
}

func TestNewSubtitlePrinter_SrtFormat(t *testing.T) {
	tests := []struct {
		name     string
		subtitle Subtitle
		want     string
	}{
		{
			name: "simple subtitle",
			subtitle: Subtitle{
				Start: 5*time.Second + 120*time.Millisecond,
				End:   6*time.Second + 840*time.Millisecond,
				Text:  "Hello, World!",
			},
			want: "1\n00:00:05,120 --> 00:00:06,840\nHello, World!\n\n",
		},
		{
			name: "subtitle with hours",
			subtitle: Subtitle{
				Start: 1*time.Hour + 23*time.Minute + 45*time.Second + 678*time.Millisecond,
				End:   1*time.Hour + 23*time.Minute + 50*time.Second + 123*time.Millisecond,
				Text:  "Long movie subtitle",
			},
			want: "1\n01:23:45,678 --> 01:23:50,123\nLong movie subtitle\n\n",
		},
		{
			name: "subtitle with multiline text",
			subtitle: Subtitle{
				Start: 10 * time.Second,
				End:   15 * time.Second,
				Text:  "First line\nSecond line\nThird line",
			},
			want: "1\n00:00:10,000 --> 00:00:15,000\nFirst line\nSecond line\nThird line\n\n",
		},
		{
			name: "subtitle at zero time",
			subtitle: Subtitle{
				Start: 0,
				End:   1*time.Second + 500*time.Millisecond,
				Text:  "Opening subtitle",
			},
			want: "1\n00:00:00,000 --> 00:00:01,500\nOpening subtitle\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printer := NewSubtitlePrinter(&buf, SrtFormat)

			if printer == nil {
				t.Fatal("expected printer function, got nil")
			}

			err := printer(tt.subtitle)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.want, got)
			}
		})
	}
}

func TestNewSubtitlePrinter_SrtFormat_MultipleSubtitles(t *testing.T) {
	var buf bytes.Buffer
	printer := NewSubtitlePrinter(&buf, SrtFormat)

	if printer == nil {
		t.Fatal("expected printer function, got nil")
	}

	subtitles := []Subtitle{
		{
			Start: 1 * time.Second,
			End:   2 * time.Second,
			Text:  "First subtitle",
		},
		{
			Start: 3 * time.Second,
			End:   4 * time.Second,
			Text:  "Second subtitle",
		},
		{
			Start: 5 * time.Second,
			End:   6 * time.Second,
			Text:  "Third subtitle",
		},
	}

	for _, sub := range subtitles {
		if err := printer(sub); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	expected := "1\n00:00:01,000 --> 00:00:02,000\nFirst subtitle\n\n" +
		"2\n00:00:03,000 --> 00:00:04,000\nSecond subtitle\n\n" +
		"3\n00:00:05,000 --> 00:00:06,000\nThird subtitle\n\n"

	got := buf.String()
	if got != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, got)
	}
}

func TestNewSubtitlePrinter_TxtFormat(t *testing.T) {
	tests := []struct {
		name     string
		subtitle Subtitle
		want     string
	}{
		{
			name: "simple subtitle",
			subtitle: Subtitle{
				Start: time.Duration(100*1001*1000/24000) * time.Millisecond,
				End:   time.Duration(200*1001*1000/24000) * time.Millisecond,
				Text:  "Hello, World!",
			},
			want: "{100}{200}Hello, World!\n",
		},
		{
			name: "subtitle with multiline text",
			subtitle: Subtitle{
				Start: time.Duration(500*1001*1000/24000) * time.Millisecond,
				End:   time.Duration(600*1001*1000/24000) * time.Millisecond,
				Text:  "First line\nSecond line\nThird line",
			},
			want: "{500}{600}First line|Second line|Third line\n",
		},
		{
			name: "subtitle at zero frame",
			subtitle: Subtitle{
				Start: 0,
				End:   time.Duration(50*1001*1000/24000) * time.Millisecond,
				Text:  "Opening subtitle",
			},
			want: "{0}{50}Opening subtitle\n",
		},
		{
			name: "subtitle with empty text",
			subtitle: Subtitle{
				Start: time.Duration(100*1001*1000/24000) * time.Millisecond,
				End:   time.Duration(150*1001*1000/24000) * time.Millisecond,
				Text:  "",
			},
			want: "{100}{150}\n",
		},
		{
			name: "subtitle with large frame numbers",
			subtitle: Subtitle{
				Start: time.Duration(100000*1001*1000/24000) * time.Millisecond,
				End:   time.Duration(200000*1001*1000/24000) * time.Millisecond,
				Text:  "Long movie subtitle",
			},
			want: "{100000}{200000}Long movie subtitle\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printer := NewSubtitlePrinter(&buf, TxtFormat)

			if printer == nil {
				t.Fatal("expected printer function, got nil")
			}

			err := printer(tt.subtitle)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.want, got)
			}
		})
	}
}

func TestNewSubtitlePrinter_TxtFormat_MultipleSubtitles(t *testing.T) {
	var buf bytes.Buffer
	printer := NewSubtitlePrinter(&buf, TxtFormat)

	if printer == nil {
		t.Fatal("expected printer function, got nil")
	}

	subtitles := []Subtitle{
		{
			Start: time.Duration(100*1001*1000/24000) * time.Millisecond,
			End:   time.Duration(150*1001*1000/24000) * time.Millisecond,
			Text:  "First subtitle",
		},
		{
			Start: time.Duration(200*1001*1000/24000) * time.Millisecond,
			End:   time.Duration(250*1001*1000/24000) * time.Millisecond,
			Text:  "Second subtitle",
		},
		{
			Start: time.Duration(300*1001*1000/24000) * time.Millisecond,
			End:   time.Duration(350*1001*1000/24000) * time.Millisecond,
			Text:  "Third subtitle",
		},
	}

	for _, sub := range subtitles {
		if err := printer(sub); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	expected := "{100}{150}First subtitle\n" +
		"{200}{250}Second subtitle\n" +
		"{300}{350}Third subtitle\n"

	got := buf.String()
	if got != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, got)
	}
}

func TestNewSubtitlePrinter_TxtFormat_WriteErrors(t *testing.T) {
	sub := Subtitle{
		Start: time.Duration(100*1001*1000/24000) * time.Millisecond,
		End:   time.Duration(200*1001*1000/24000) * time.Millisecond,
		Text:  "Test subtitle",
	}

	tests := []struct {
		name      string
		failAfter int
	}{
		{
			name:      "error writing start duration",
			failAfter: 0,
		},
		{
			name:      "error writing end duration",
			failAfter: 1,
		},
		{
			name:      "error writing text",
			failAfter: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &errorWriter{failAfter: tt.failAfter}
			printer := NewSubtitlePrinter(writer, TxtFormat)

			if printer == nil {
				t.Fatal("expected printer function, got nil")
			}

			err := printer(sub)
			if err == nil {
				t.Error("expected error from writer, got nil")
			}

			if !strings.Contains(err.Error(), "write error") {
				t.Errorf("expected 'write error', got: %v", err)
			}
		})
	}
}

func TestNewSubtitlePrinter_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	printer := NewSubtitlePrinter(&buf, UnknownFormat)

	if printer != nil {
		t.Errorf("expected nil for UnknownFormat, got function")
	}
}

// errorWriter is a writer that always returns an error
type errorWriter struct {
	writeCount int
	failAfter  int
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	w.writeCount++
	if w.writeCount > w.failAfter {
		return 0, errors.New("write error")
	}
	return len(p), nil
}

func TestNewSubtitlePrinter_SrtFormat_WriteErrors(t *testing.T) {
	sub := Subtitle{
		Start: 1 * time.Second,
		End:   2 * time.Second,
		Text:  "Test subtitle",
	}

	tests := []struct {
		name      string
		failAfter int
	}{
		{
			name:      "error writing subtitle number",
			failAfter: 0,
		},
		{
			name:      "error writing start duration",
			failAfter: 1,
		},
		{
			name:      "error writing arrow separator",
			failAfter: 2,
		},
		{
			name:      "error writing end duration",
			failAfter: 3,
		},
		{
			name:      "error writing text",
			failAfter: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &errorWriter{failAfter: tt.failAfter}
			printer := NewSubtitlePrinter(writer, SrtFormat)

			if printer == nil {
				t.Fatal("expected printer function, got nil")
			}

			err := printer(sub)
			if err == nil {
				t.Error("expected error from writer, got nil")
			}

			if !strings.Contains(err.Error(), "write error") {
				t.Errorf("expected 'write error', got: %v", err)
			}
		})
	}
}
