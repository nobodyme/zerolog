package zerolog_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func ExampleConsoleWriter() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true})

	log.Info().Str("foo", "bar").Msg("Hello World")
	// Output: <nil> INF Hello World foo=bar
}

func ExampleConsoleWriter_customFormatters() {
	out := zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true}
	out.FormatLevel = func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("%-6s|", i)) }
	out.FormatFieldName = func(i interface{}) string { return fmt.Sprintf("%s:", i) }
	out.FormatFieldValue = func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("%s", i)) }
	log := zerolog.New(out)

	log.Info().Str("foo", "bar").Msg("Hello World")
	// Output: <nil> INFO  | Hello World foo:BAR
}

func ExampleNewConsoleWriter() {
	out := zerolog.NewConsoleWriter()
	out.NoColor = true // For testing purposes only
	log := zerolog.New(out)

	log.Debug().Str("foo", "bar").Msg("Hello World")
	// Output: <nil> DBG Hello World foo=bar
}

func ExampleNewConsoleWriter_customFormatters() {
	out := zerolog.NewConsoleWriter(
		func(w *zerolog.ConsoleWriter) {
			// Customize time format
			w.TimeFormat = time.RFC822
			// Customize level formatting
			w.FormatLevel = func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("[%-5s]", i)) }
		},
	)
	out.NoColor = true // For testing purposes only

	log := zerolog.New(out)

	log.Info().Str("foo", "bar").Msg("Hello World")
	// Output: <nil> [INFO ] Hello World foo=bar
}

func TestConsoleLogger(t *testing.T) {
	t.Run("Numbers", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := zerolog.New(zerolog.ConsoleWriter{Out: buf, NoColor: true})
		log.Info().
			Float64("float", 1.23).
			Uint64("small", 123).
			Uint64("big", 1152921504606846976).
			Msg("msg")
		if got, want := strings.TrimSpace(buf.String()), "<nil> INF msg big=1152921504606846976 float=1.23 small=123"; got != want {
			t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
		}
	})
}

func TestConsoleWriter(t *testing.T) {
	t.Run("Default field formatter", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true, PartsOrder: []string{"foo"}}

		_, err := w.Write([]byte(`{"foo": "DEFAULT"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "DEFAULT foo=DEFAULT\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Write colorized", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: false}

		_, err := w.Write([]byte(`{"level": "warn", "message": "Foobar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "\x1b[90m<nil>\x1b[0m \x1b[33mWRN\x1b[0m \x1b[1mFoobar\x1b[0m\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("NO_COLOR = true", func(t *testing.T) {
		os.Setenv("NO_COLOR", "anything")

		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf}

		_, err := w.Write([]byte(`{"level": "warn", "message": "Foobar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "<nil> WRN Foobar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
		os.Unsetenv("NO_COLOR")
	})

	t.Run("Write fields", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true}

		ts := time.Unix(0, 0)
		d := ts.UTC().Format(time.RFC3339)
		_, err := w.Write([]byte(`{"time": "` + d + `", "level": "debug", "message": "Foobar", "foo": "bar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := ts.Format(time.Kitchen) + " DBG Foobar foo=bar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Unix timestamp input format", func(t *testing.T) {
		of := zerolog.TimeFieldFormat
		defer func() {
			zerolog.TimeFieldFormat = of
		}()
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, TimeFormat: time.StampMilli, NoColor: true}

		_, err := w.Write([]byte(`{"time": 1234, "level": "debug", "message": "Foobar", "foo": "bar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := time.Unix(1234, 0).Format(time.StampMilli) + " DBG Foobar foo=bar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Unix timestamp ms input format", func(t *testing.T) {
		of := zerolog.TimeFieldFormat
		defer func() {
			zerolog.TimeFieldFormat = of
		}()
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, TimeFormat: time.StampMilli, NoColor: true}

		_, err := w.Write([]byte(`{"time": 1234567, "level": "debug", "message": "Foobar", "foo": "bar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := time.Unix(1234, 567000000).Format(time.StampMilli) + " DBG Foobar foo=bar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Unix timestamp us input format", func(t *testing.T) {
		of := zerolog.TimeFieldFormat
		defer func() {
			zerolog.TimeFieldFormat = of
		}()
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro

		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, TimeFormat: time.StampMicro, NoColor: true}

		_, err := w.Write([]byte(`{"time": 1234567891, "level": "debug", "message": "Foobar", "foo": "bar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := time.Unix(1234, 567891000).Format(time.StampMicro) + " DBG Foobar foo=bar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("No message field", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true}

		_, err := w.Write([]byte(`{"level": "debug", "foo": "bar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "<nil> DBG foo=bar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("No level field", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true}

		_, err := w.Write([]byte(`{"message": "Foobar", "foo": "bar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "<nil> ??? Foobar foo=bar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Write colorized fields", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: false}

		_, err := w.Write([]byte(`{"level": "warn", "message": "Foobar", "foo": "bar"}`))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "\x1b[90m<nil>\x1b[0m \x1b[33mWRN\x1b[0m \x1b[1mFoobar\x1b[0m \x1b[36mfoo=\x1b[0mbar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Write error field", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true}

		ts := time.Unix(0, 0)
		d := ts.UTC().Format(time.RFC3339)
		evt := `{"time": "` + d + `", "level": "error", "message": "Foobar", "aaa": "bbb", "error": "Error"}`
		// t.Log(evt)

		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := ts.Format(time.Kitchen) + " ERR Foobar error=Error aaa=bbb\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Write caller field", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true}

		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Cannot get working directory: %s", err)
		}

		ts := time.Unix(0, 0)
		d := ts.UTC().Format(time.RFC3339)
		evt := `{"time": "` + d + `", "level": "debug", "message": "Foobar", "foo": "bar", "caller": "` + cwd + `/foo/bar.go"}`
		// t.Log(evt)

		_, err = w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := ts.Format(time.Kitchen) + " DBG foo/bar.go > Foobar foo=bar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Write JSON field", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true}

		evt := `{"level": "debug", "message": "Foobar", "foo": [1, 2, 3], "bar": true}`
		// t.Log(evt)

		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "<nil> DBG Foobar bar=true foo=[1,2,3]\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})
}

func TestConsoleWriterConfiguration(t *testing.T) {
	t.Run("Sets TimeFormat", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true, TimeFormat: time.RFC3339}

		ts := time.Unix(0, 0)
		d := ts.UTC().Format(time.RFC3339)
		evt := `{"time": "` + d + `", "level": "info", "message": "Foobar"}`

		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := ts.Format(time.RFC3339) + " INF Foobar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Sets PartsOrder", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true, PartsOrder: []string{"message", "level"}}

		evt := `{"level": "info", "message": "Foobar"}`
		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "Foobar INF\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Sets PartsExclude", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true, PartsExclude: []string{"time"}}

		d := time.Unix(0, 0).UTC().Format(time.RFC3339)
		evt := `{"time": "` + d + `", "level": "info", "message": "Foobar"}`
		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "INF Foobar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Sets FieldsExclude", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true, FieldsExclude: []string{"foo"}}

		evt := `{"level": "info", "message": "Foobar", "foo":"bar", "baz":"quux"}`
		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "<nil> INF Foobar baz=quux\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Sets FormatExtra", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{
			Out: buf, NoColor: true, PartsOrder: []string{"level", "message"},
			FormatExtra: func(evt map[string]interface{}, buf *bytes.Buffer) error {
				buf.WriteString("\nAdditional stacktrace")
				return nil
			},
		}

		evt := `{"level": "info", "message": "Foobar"}`
		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "INF Foobar\nAdditional stacktrace\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Sets FormatPrepare", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{
			Out: buf, NoColor: true, PartsOrder: []string{"level", "message"},
			FormatPrepare: func(evt map[string]interface{}) error {
				evt["message"] = fmt.Sprintf("msg=%s", evt["message"])
				return nil
			},
		}

		evt := `{"level": "info", "message": "Foobar"}`
		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		expectedOutput := "INF msg=Foobar\n"
		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})

	t.Run("Uses local time for console writer without time zone", func(t *testing.T) {
		// Regression test for issue #483 (check there for more details)

		timeFormat := "2006-01-02 15:04:05"
		expectedOutput := "2022-10-20 20:24:50 INF Foobar\n"
		evt := `{"time": "2022-10-20 20:24:50", "level": "info", "message": "Foobar"}`

		of := zerolog.TimeFieldFormat
		defer func() {
			zerolog.TimeFieldFormat = of
		}()
		zerolog.TimeFieldFormat = timeFormat

		buf := &bytes.Buffer{}
		w := zerolog.ConsoleWriter{Out: buf, NoColor: true, TimeFormat: timeFormat}
		_, err := w.Write([]byte(evt))
		if err != nil {
			t.Errorf("Unexpected error when writing output: %s", err)
		}

		actualOutput := buf.String()
		if actualOutput != expectedOutput {
			t.Errorf("Unexpected output %q, want: %q", actualOutput, expectedOutput)
		}
	})
}

func BenchmarkConsoleWriter(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	var msg = []byte(`{"level": "info", "foo": "bar", "message": "HELLO", "time": "1990-01-01"}`)

	w := zerolog.ConsoleWriter{Out: io.Discard, NoColor: false}

	for i := 0; i < b.N; i++ {
		w.Write(msg)
	}
}
