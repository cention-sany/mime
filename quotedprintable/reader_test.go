// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quotedprintable

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestStrictReader(t *testing.T) {
	tests := []struct {
		in, want string
		err      interface{}
	}{
		{in: "", want: ""},
		{in: "foo bar", want: "foo bar"},
		{in: "foo bar=3D", want: "foo bar="},
		{in: "foo bar=3d", want: "foo bar="}, // lax.
		{in: "foo bar=\n", want: "foo bar"},
		{in: "foo bar\n", want: "foo bar\n"}, // somewhat lax.
		{in: "foo bar=0", want: "foo bar=0", err: io.ErrUnexpectedEOF},
		{in: "foo bar=0D=0A", want: "foo bar\r\n"},
		{in: " A B        \r\n C ", want: " A B\r\n C"},
		{in: " A B =\r\n C ", want: " A B  C"},
		{in: " A B =\n C ", want: " A B  C"}, // lax. treating LF as CRLF
		{in: "foo=\nbar", want: "foobar"},
		{in: "foo\x00bar", want: "foo\x00bar", err: "quotedprintable: invalid unescaped byte 0x00 in body"},
		{in: "foo bar\xff", want: "foo bar\xff", err: "quotedprintable: invalid unescaped byte 0xff in body"},

		// Equal sign.
		{in: "=3D30\n", want: "=30\n"},
		{in: "=00=FF0=\n", want: "\x00\xff0"},

		// Trailing whitespace
		{in: "foo  \n", want: "foo\n"},
		{in: "foo  \n\nfoo =\n\nfoo=20\n\n", want: "foo\n\nfoo \nfoo \n\n"},

		// Tests that we allow bare \n and \r through, despite it being strictly
		// not permitted per RFC 2045, Section 6.7 Page 22 bullet (4).
		{in: "foo\nbar", want: "foo\nbar"},
		{in: "foo\rbar", want: "foo\rbar"},
		{in: "foo\r\nbar", want: "foo\r\nbar"},
		{in: "foo=\r\nbar", want: "foobar"},

		// Different types of soft line-breaks.
		{in: "foo=\r\nbar", want: "foobar"},
		{in: "foo=\nbar", want: "foobar"},
		{in: "foo=\rbar", want: "foo=\rbar", err: "quotedprintable: invalid hex byte 0x0d"},
		{in: "foo=\r\r\r \nbar", want: "foo", err: `quotedprintable: invalid bytes after =: "\r\r\r \n"`},

		// Example from RFC 2045:
		{in: "Now's the time =\n" + "for all folk to come=\n" + " to the aid of their country.",
			want: "Now's the time for all folk to come to the aid of their country."},

		// Cention bad email
		{in: "Sendt fra min iPad=", want: "Sendt fra min iPad", err: "quotedprintable: invalid bytes after =: \"\""},
		{in: "<div src=\"http://123.456.789.88\">", want: "<div src=\"http://123.456.789.88\">",
			err: "quotedprintable: invalid hex byte 0x22"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, NewStrictReader(strings.NewReader(tt.in)))
		if got := buf.String(); got != tt.want {
			t.Errorf("for %q, got %q; want %q", tt.in, got, tt.want)
		}
		switch verr := tt.err.(type) {
		case nil:
			if err != nil && err != io.EOF {
				t.Errorf("for %q, got unexpected error: %v", tt.in, err)
			}
		case string:
			if got := fmt.Sprint(err); got != verr {
				t.Errorf("for %q, got error %q; want %q", tt.in, got, verr)
			}
		case error:
			if err != verr {
				t.Errorf("for %q, got error %q; want %q", tt.in, err, verr)
			}
		}
	}
}

func TestReader(t *testing.T) {
	tests := []struct {
		in, want string
		err      interface{}
	}{
		{in: "", want: ""},
		{in: "foo bar", want: "foo bar"},
		{in: "foo bar=3D", want: "foo bar="},
		{in: "foo bar=3d", want: "foo bar="}, // lax.
		{in: "foo bar=\n", want: "foo bar"},
		{in: "foo bar\n", want: "foo bar\n"}, // somewhat lax.
		{in: "foo bar=0", want: "foo bar=0"},
		{in: "foo bar=0D=0A", want: "foo bar\r\n"},
		{in: " A B        \r\n C ", want: " A B\r\n C"},
		{in: " A B =\r\n C ", want: " A B  C"},
		{in: " A B =\n C ", want: " A B  C"}, // lax. treating LF as CRLF
		{in: "foo=\nbar", want: "foobar"},
		{in: "foo\x00bar", want: "foo\x00bar"},
		{in: "foo bar\xff", want: "foo bar\xff"},

		// Equal sign.
		{in: "=3D30\n", want: "=30\n"},
		{in: "=00=FF0=\n", want: "\x00\xff0"},

		// Trailing whitespace
		{in: "foo  \n", want: "foo\n"},
		{in: "foo  \n\nfoo =\n\nfoo=20\n\n", want: "foo\n\nfoo \nfoo \n\n"},

		// Tests that we allow bare \n and \r through, despite it being strictly
		// not permitted per RFC 2045, Section 6.7 Page 22 bullet (4).
		{in: "foo\nbar", want: "foo\nbar"},
		{in: "foo\rbar", want: "foo\rbar"},
		{in: "foo\r\nbar", want: "foo\r\nbar"},
		{in: "foo=\r\nbar", want: "foobar"},

		// Different types of soft line-breaks.
		{in: "foo=\r\nbar", want: "foobar"},
		{in: "foo=\nbar", want: "foobar"},
		{in: "foo=\rbar", want: "foo=\rbar"},
		{in: "foo=\r\r\r \nbar", want: "foobar"},

		// Example from RFC 2045:
		{in: "Now's the time =\n" + "for all folk to come=\n" + " to the aid of their country.",
			want: "Now's the time for all folk to come to the aid of their country."},

		// Cention bad email
		{in: "Sendt fra min iPad=", want: "Sendt fra min iPad"},
		{in: "<div src=\"http://123.456.789.88\">", want: "<div src=\"http://123.456.789.88\">"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, NewReader(strings.NewReader(tt.in)))
		if got := buf.String(); got != tt.want {
			t.Errorf("for %q, got %q; want %q", tt.in, got, tt.want)
		}
		switch verr := tt.err.(type) {
		case nil:
			if err != nil {
				t.Errorf("for %q, got unexpected error: %v", tt.in, err)
			}
		case string:
			if got := fmt.Sprint(err); got != verr {
				t.Errorf("for %q, got error %q; want %q", tt.in, got, verr)
			}
		case error:
			if err != verr {
				t.Errorf("for %q, got error %q; want %q", tt.in, err, verr)
			}
		}
	}
}

func everySequence(base, alpha string, length int, fn func(string)) {
	if len(base) == length {
		fn(base)
		return
	}
	for i := 0; i < len(alpha); i++ {
		everySequence(base+alpha[i:i+1], alpha, length, fn)
	}
}

var useQprint = flag.Bool("qprint", false, "Compare against the 'qprint' program.")

var badSoftRx = regexp.MustCompile(`=([^\r\n]+?\n)|([^\r\n]+$)|(\r$)|(\r[^\n]+\n)|( \r\n)`)

func TestExhaustive(t *testing.T) {
	if *useQprint {
		_, err := exec.LookPath("qprint")
		if err != nil {
			t.Fatalf("Error looking for qprint: %v", err)
		}
	}

	var buf bytes.Buffer
	res := make(map[string]int)
	everySequence("", "0A \r\n=", 6, func(s string) {
		if strings.HasSuffix(s, "=") || strings.Contains(s, "==") {
			return
		}
		buf.Reset()
		_, err := io.Copy(&buf, NewStrictReader(strings.NewReader(s)))
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "invalid bytes after =:") {
				errStr = "invalid bytes after ="
			}
			res[errStr]++
			if strings.Contains(errStr, "invalid hex byte ") {
				if strings.HasSuffix(errStr, "0x20") && (strings.Contains(s, "=0 ") || strings.Contains(s, "=A ") || strings.Contains(s, "= ")) {
					return
				}
				if strings.HasSuffix(errStr, "0x3d") && (strings.Contains(s, "=0=") || strings.Contains(s, "=A=")) {
					return
				}
				if strings.HasSuffix(errStr, "0x0a") || strings.HasSuffix(errStr, "0x0d") {
					// bunch of cases; since whitespace at the end of a line before \n is removed.
					return
				}
			}
			if strings.Contains(errStr, "unexpected EOF") {
				return
			}
			if errStr == "invalid bytes after =" && badSoftRx.MatchString(s) {
				return
			}
			t.Errorf("decode(%q) = %v", s, err)
			return
		}
		if *useQprint {
			cmd := exec.Command("qprint", "-d")
			cmd.Stdin = strings.NewReader(s)
			stderr, err := cmd.StderrPipe()
			if err != nil {
				panic(err)
			}
			qpres := make(chan interface{}, 2)
			go func() {
				br := bufio.NewReader(stderr)
				s, _ := br.ReadString('\n')
				if s != "" {
					qpres <- errors.New(s)
					if cmd.Process != nil {
						// It can get stuck on invalid input, like:
						// echo -n "0000= " | qprint -d
						cmd.Process.Kill()
					}
				}
			}()
			go func() {
				want, err := cmd.Output()
				if err == nil {
					qpres <- want
				}
			}()
			select {
			case got := <-qpres:
				if want, ok := got.([]byte); ok {
					if string(want) != buf.String() {
						t.Errorf("go decode(%q) = %q; qprint = %q", s, want, buf.String())
					}
				} else {
					t.Logf("qprint -d(%q) = %v", s, got)
				}
			case <-time.After(5 * time.Second):
				t.Logf("qprint timeout on %q", s)
			}
		}
		res["OK"]++
	})
	var outcomes []string
	for k, v := range res {
		outcomes = append(outcomes, fmt.Sprintf("%v: %d", k, v))
	}
	sort.Strings(outcomes)
	got := strings.Join(outcomes, "\n")
	want := `OK: 21576
invalid bytes after =: 4081
quotedprintable: invalid hex byte 0x0a: 1400
quotedprintable: invalid hex byte 0x0d: 2554
quotedprintable: invalid hex byte 0x20: 2344
quotedprintable: invalid hex byte 0x3d: 424
unexpected EOF: 2746`
	// 	want := `OK: 21576
	// invalid bytes after =: 3397
	// quotedprintable: invalid hex byte 0x0a: 1400
	// quotedprintable: invalid hex byte 0x0d: 2700
	// quotedprintable: invalid hex byte 0x20: 2490
	// quotedprintable: invalid hex byte 0x3d: 440
	// unexpected EOF: 3122`
	if got != want {
		t.Errorf("Got:\n%s\nWant:\n%s", got, want)
	}
}

func TestLongBuffer(t *testing.T) {
	const (
		testLongSize = 4100
	)
	var (
		tstdata [testLongSize]byte
		data2   [testLongSize + 4]byte
		buf     bytes.Buffer
	)
	for i := 0; i < testLongSize; i++ {
		tstdata[i] = 'A'
	}
	_, err := io.Copy(&buf, NewReader(strings.NewReader(string(tstdata[:]))))
	if err != nil {
		t.Errorf("Expect nil error but got: %v", err)
	}
	want := string(tstdata[:])
	got := buf.String()
	if got != want {
		t.Errorf("Test identity: Expect %s but got %s", want, got)
	}

	buf.Reset()
	for i, v := range tstdata {
		data2[i] = v
	}
	data2[4095] = '='
	data2[4096] = '\n'
	_, err = io.Copy(&buf, NewReader(strings.NewReader(string(data2[:testLongSize]))))
	if err != nil {
		t.Errorf("Expect nil error but got: %v", err)
	}
	want = string(tstdata[:testLongSize-2])
	got = buf.String()
	if got != want {
		t.Errorf("Test break LF: Expect %s but got %s", want, got)
	}

	buf.Reset()
	for i, v := range tstdata {
		data2[i] = v
	}
	data2[4095] = '='
	data2[4096] = '\r'
	data2[4097] = '\n'
	_, err = io.Copy(&buf, NewReader(strings.NewReader(string(data2[:testLongSize]))))
	if err != nil {
		t.Errorf("Expect nil error but got: %v", err)
	}
	want = string(tstdata[:testLongSize-3])
	got = buf.String()
	if got != want {
		t.Errorf("Test break CRLF: Expect %s but got %s", want, got)
	}

	buf.Reset()
	for i, v := range tstdata {
		data2[i] = v
	}
	data2[4095] = '='
	data2[4096] = '{'
	data2[4097] = '}'
	_, err = io.Copy(&buf, NewReader(strings.NewReader(string(data2[:testLongSize]))))
	if err != nil {
		t.Errorf("Expect nil error but got: %v", err)
	}
	want = string(data2[:testLongSize])
	got = buf.String()
	if got != want {
		t.Errorf("Test unescape equal 1: Expect %s but got %s", want, got)
	}

	buf.Reset()
	for i, v := range tstdata {
		data2[i] = v
	}
	data2[4095] = '='
	data2[4096] = '{'
	_, err = io.Copy(&buf, NewReader(strings.NewReader(string(data2[:4097]))))
	if err != nil {
		t.Errorf("Expect nil error but got: %v", err)
	}
	want = string(data2[:4097])
	got = buf.String()
	if got != want {
		t.Errorf("Test unescape equal 2: Expect %s but got %s", want, got)
	}

	buf.Reset()
	for i, v := range tstdata {
		data2[i] = v
	}
	data2[4094] = '='
	data2[4095] = '{'
	data2[4096] = '}'
	_, err = io.Copy(&buf, NewReader(strings.NewReader(string(data2[:testLongSize]))))
	if err != nil {
		t.Errorf("Expect nil error but got: %v", err)
	}
	want = string(data2[:testLongSize])
	got = buf.String()
	if got != want {
		t.Errorf("Test unescape equal 3: Expect %s but got %s", want, got)
	}

	buf.Reset()
	for i := range data2 {
		data2[i] = '='
	}
	_, err = io.Copy(&buf, NewReader(strings.NewReader(string(data2[:]))))
	if err != nil {
		t.Errorf("Expect nil error but got: %v", err)
	}
	// last '=' b4 io.EOF always get truncate because it is assumed as
	// quoted-printable soft break.
	want = string(data2[:len(data2)-1])
	got = buf.String()
	if got != want {
		t.Errorf("Test all equal sign: Expect %s but got %s", want, got)
	}

	buf.Reset()
	want = string(tstdata[:4095]) + "☎" // black telephone \u260e
	for i := 0; i < 4095; i++ {
		data2[i] = tstdata[i]
	}
	data2[4095] = '='
	data2[4096] = 'E'
	data2[4097] = '2'
	data2[4098] = '='
	data2[4099] = '9'
	data2[4100] = '8'
	data2[4101] = '='
	data2[4102] = '8'
	data2[4103] = 'E'
	_, err = io.Copy(&buf, NewReader(strings.NewReader(string(data2[:]))))
	if err != nil {
		t.Errorf("Expect nil error but got: %v", err)
	}
	got = buf.String()
	if got != want {
		t.Errorf("Test break non-ascii: Expect %s but got %s", want, got)
	}
}

func TestErrorReader(t *testing.T) {
	rerr := new(RErr)
	const noError = "<nil>"
	if s := rerr.Error(); s != noError {
		t.Errorf("Expect empty error return %s but got: %s", noError, s)
	}
	rerr.add(nil)
	rerr.add(errors.New("test1"))
	rerr.add(errors.New("test2"))
	const multiErrors = "test1|test2"
	if s := rerr.Error(); s != multiErrors {
		t.Errorf("Expect error string: %s but got: %s", multiErrors, s)
	}
	const badError = "test3"
	rerr.addUnrecover(errors.New(badError))
	if s := rerr.Error(); s != badError {
		t.Errorf("Expect error string: %s but got: %s", badError, s)
	}
}
