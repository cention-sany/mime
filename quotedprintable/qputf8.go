// quotedprintable utf8 specifically correct expected utf8 and quotedprintable.

package quotedprintable

import (
	"io"
	"unicode/utf8"
)

const (
	stNormal = iota
	stStart
	st0x0D
	st0x0A
	st0x10XXXXXX
	stRelease
	stReleaseRestart
)

type qpUTF8 struct {
	r     io.Reader
	state int
	pos   int // index of first end-of-line char on own buffer
	pco   int // producer counter for own buffer
	cco   int // consumer counter for own buffer
	own   [6]byte
	pcb   int // producer counter for buf buffer
	ccb   int // consumer counter for buf buffer
	buf   [512]byte
	last  byte
	err   error
}

func newQPUTF8(r io.Reader) *qpUTF8 {
	return &qpUTF8{r: r}
}

// Read implements io.Reader interface.
func (q *qpUTF8) Read(p []byte) (int, error) {
	var count int
	max := len(p)
	if q.pcb > q.ccb {
		if !q.cycle(p, q.ccb, q.pcb, max, &count) {
			return count, nil
		}
	}
	if q.err != nil {
		if q.state == st0x10XXXXXX || (q.state == stStart && q.pco-q.cco > 1) {
			q.release(p, &count, max, false)
		}
		return count, q.err
	}
	var n int
	allowedReadSize := (max - count) - (q.pco - q.cco)
	if allowedReadSize > 0 {
		if allowedReadSize > 512 {
			n, q.err = q.r.Read(q.buf[:])
		} else {
			n, q.err = q.r.Read(q.buf[:allowedReadSize])
		}
	}
	if q.cycle(p, 0, n, max, &count) {
		q.pcb = 0
		q.ccb = 0
		if q.cco >= q.pco {
			return count, q.err
		}
	}
	return count, nil
}

func (q *qpUTF8) cycle(p []byte, start, n, max int, count *int) bool {
	for i := start; i < n; i++ {
		b := q.buf[i]
		switch q.state {
		case stNormal:
			if b&0xc0 == 0xc0 {
				q.state = stStart
				q.own[0] = b
				q.pco = 1
			} else {
				p[*count] = b
				*count++
				if *count >= max {
					q.pcb = n
					q.ccb = i + 1
					return false
				}
			}
		case stStart:
			c0mask := b & 0xc0
			if c0mask == 0xc0 {
				q.last = b
				q.state = stReleaseRestart
				break
			}
			currentPos := q.pco
			q.own[q.pco] = b
			q.pco++
			if b == 0x0d {
				q.pos = currentPos
				q.state = st0x0D
			} else if b == 0x0a {
				q.pos = currentPos
				q.state = st0x0A
			} else if c0mask != 0x80 || q.pco >= utf8.UTFMax {
				q.state = stRelease
			}
		case st0x0D:
			if b&0xc0 == 0xc0 {
				q.last = b
				q.state = stReleaseRestart
				break
			}
			q.own[q.pco] = b
			q.pco++
			if b == 0x0a {
				q.state = st0x0A
			} else {
				q.state = stRelease
			}
		case st0x0A:
			c0mask := b & 0xc0
			if c0mask == 0xc0 {
				q.last = b
				q.state = stReleaseRestart
			} else if c0mask == 0x80 {
				q.own[q.pos] = b
				q.pco = q.pos + 1
				q.state = st0x10XXXXXX
			} else {
				q.own[q.pco] = b
				q.pco++
				q.state = stRelease
			}
		case st0x10XXXXXX:
			c0mask := b & 0xc0
			if c0mask == 0xc0 {
				q.last = b
				q.state = stReleaseRestart
				break
			}
			q.own[q.pco] = b
			q.pco++
			if c0mask != 0x80 || q.pco == 6 {
				q.state = stRelease
			}
		}
		isReleaseRestart := q.state == stReleaseRestart
		if isReleaseRestart || q.state == stRelease {
			if !q.release(p, count, max, isReleaseRestart) {
				q.pcb = n
				q.ccb = i + 1
				return false
			}
		}
	}
	return true
}

func (q *qpUTF8) release(p []byte, count *int, max int, restart bool) bool {
	for {
		if q.cco < q.pco {
			p[*count] = q.own[q.cco]
			q.cco++
			*count++
			if *count >= max {
				return false
			}
		} else {
			// reset
			q.cco = 0
			if restart {
				q.own[0] = q.last
				q.pco = 1
				q.state = stStart
			} else {
				q.pco = 0
				q.state = stNormal
			}
			break
		}
	}
	return true
}

func NewUTF8Reader(r io.Reader) io.Reader {
	return newQPUTF8(NewReader(r))
}
