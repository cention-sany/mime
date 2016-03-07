// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package quotedprintable implements quoted-printable encoding as specified by
// RFC 2045.
package quotedprintable

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const maxQPRErr = 4

type RErr struct {
	count int
	errs  [maxQPRErr]error // record errors that can be ignored
	err   error            // special field for serious error or non-ignorable
}

func (q *RErr) Error() string {
	if q.err != nil {
		return q.err.Error()
	} else if q.count > 0 {
		return q.errs[0].Error()
	}
	return ""
}

// add error that is recoverable. After this, the modified func
// could still continue processing amid of this mild err error.
func (q *RErr) add(err error) {
	if err == nil {
		return
	}
	if q.count == maxQPRErr {
		return
	}
	q.errs[q.count] = err
	q.count++
}

// add error that is not recoverable. After this, there should not
// have any more processing.
func (q *RErr) addUnrecover(err error) {
	if err != nil {
		q.err = err
	}
}

// return the error that almost same as stdlib - maintain compatibility
func (q *RErr) getErr() error {
	if q.count == 0 {
		return q.err
	}
	return q.errs[0]
}

// Reader is a quoted-printable decoder.
type Reader struct {
	Fn   func() error
	br   *bufio.Reader
	gerr *RErr
	rerr error  // last read error
	line []byte // to be consumed before more of br
}

func getReader(r io.Reader) *Reader {
	return &Reader{
		br:   bufio.NewReader(r),
		gerr: new(RErr),
	}
}

// NewReader returns a quoted-printable reader, decoding from r. It return
// error as similar to stdlib as possible.
func NewStrictReader(r io.Reader) *Reader {
	rd := getReader(r)
	rd.Fn = func() error {
		return rd.gerr.getErr()
	}
	return rd
}

// NewReader returns a quoted-printable reader, decoding from r. It only
// return unrecoverable error or EOF as non-nil.
func NewReader(r io.Reader) *Reader {
	rd := getReader(r)
	rd.Fn = func() error {
		return rd.gerr.err
	}
	return rd
}

func fromHex(b byte) (byte, error) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', nil
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, nil
	// Accept badly encoded bytes.
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, nil
	}
	return 0, fmt.Errorf("quotedprintable: invalid hex byte 0x%02x", b)
}

func readHexByte(v []byte) (b byte, err error) {
	if len(v) < 2 {
		return 0, io.ErrUnexpectedEOF
	}
	var hb, lb byte
	if hb, err = fromHex(v[0]); err != nil {
		return 0, err
	}
	if lb, err = fromHex(v[1]); err != nil {
		return 0, err
	}
	return hb<<4 | lb, nil
}

func isQPDiscardWhitespace(r rune) bool {
	switch r {
	case '\n', '\r', ' ', '\t':
		return true
	}
	return false
}

var (
	crlf       = []byte("\r\n")
	lf         = []byte("\n")
	softSuffix = []byte("=")
)

// Read reads and decodes quoted-printable data from the underlying reader.
func (r *Reader) Read(p []byte) (int, error) {
	// Deviations from RFC 2045:
	// 1. in addition to "=\r\n", "=\n" is also treated as soft line break.
	// 2. it will pass through a '\r' or '\n' not preceded by '=', consistent
	//    with other broken QP encoders & decoders.
	var n int
	var err error
	for len(p) > 0 {
		if len(r.line) == 0 {
			if err = r.Fn(); err != nil {
				return n, err
			}
			r.line, r.rerr = r.br.ReadSlice('\n')
			r.gerr.addUnrecover(r.rerr)

			// Does the line end in CRLF instead of just LF?
			hasLF := bytes.HasSuffix(r.line, lf)
			hasCR := bytes.HasSuffix(r.line, crlf)
			wholeLine := r.line
			r.line = bytes.TrimRightFunc(wholeLine, isQPDiscardWhitespace)
			if bytes.HasSuffix(r.line, softSuffix) {
				rightStripped := wholeLine[len(r.line):]
				r.line = r.line[:len(r.line)-1]
				if !bytes.HasPrefix(rightStripped, lf) && !bytes.HasPrefix(rightStripped, crlf) {
					r.rerr = fmt.Errorf("quotedprintable: invalid bytes after =: %q", rightStripped)
					r.gerr.add(r.rerr)
				}
			} else if hasLF {
				if hasCR {
					r.line = append(r.line, '\r', '\n')
				} else {
					r.line = append(r.line, '\n')
				}
			}
			continue
		}
		b := r.line[0]

		switch {
		case b == '=':
			b, err = readHexByte(r.line[1:])
			if err != nil {
				b = '='
				r.gerr.add(err)
				break // this modification allow bad email to be parsed too
				//return n, err
			}
			r.line = r.line[2:] // 2 of the 3; other 1 is done below
		case b == '\t' || b == '\r' || b == '\n':
		case b < ' ' || b > '~':
			//return n, fmt.Errorf("quotedprintable: invalid unescaped byte 0x%02x in body", b)
			r.gerr.add(fmt.Errorf("quotedprintable: invalid unescaped byte 0x%02x in body", b))
		}
		p[0] = b
		p = p[1:]
		r.line = r.line[1:]
		n++
	}
	return n, r.Fn()
}
