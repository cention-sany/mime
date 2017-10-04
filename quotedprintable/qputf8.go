// quotedprintable utf8 specifically correct expected utf8 and quotedprintable.

package quotedprintable

import (
	"io"
)

const (
	stNormal = iota
	stStart
	st0x0D
	st0x0A
	st0x10XXXXXX
	stRelease
)

type qpUTF8 struct {
	r     io.Reader
	state int
	pco   int // producer counter for own buffer
	cco   int // consumer counter for own buffer
	own   [6]byte
	pcb   int // producer counter for buf buffer
	ccb   int // consumer counter for buf buffer
	buf   [512]byte
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
	var (
		n   int
		err error
	)
	allowedReadSize := (max - count) - (q.pco - q.cco)
	if allowedReadSize > 0 {
		if allowedReadSize > 512 {
			n, err = q.r.Read(q.buf[:])
		} else {
			n, err = q.r.Read(q.buf[:allowedReadSize])
		}
	}
	if q.cycle(p, 0, n, max, &count) {
		q.pcb = 0
		q.ccb = 0
	}
	return count, err
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
			q.own[q.pco] = b
			q.pco++
			if b == 0x0d {
				q.state = st0x0D
			} else if b == 0x0a {
				q.state = st0x0A
			} else {
				q.state = stRelease
			}
		case st0x0D:
			q.own[q.pco] = b
			q.pco++
			if b == 0x0a {
				q.state = st0x0A
			} else {
				q.state = stRelease
			}
		case st0x0A:
			if b&0xc0 == 0x80 {
				q.own[1] = b
				q.pco = 2
				q.state = st0x10XXXXXX
			} else {
				q.own[q.pco] = b
				q.pco++
				q.state = stRelease
			}
		case st0x10XXXXXX:
			q.own[q.pco] = b
			q.pco++
			if b&0xc0 != 0x80 || q.pco == 6 {
				q.state = stRelease
			}
		}
		if q.state == stRelease {
			if !q.release(p, count, max) {
				q.pcb = n
				q.ccb = i + 1
				return false
			}
		}
	}
	return true
}

func (q *qpUTF8) release(p []byte, count *int, max int) bool {
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
			q.pco = 0
			q.state = stNormal
			break
		}
	}
	return true
}

func NewUTF8Reader(r io.Reader) io.Reader {
	return newQPUTF8(NewReader(r))
}
