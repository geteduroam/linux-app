package discovery

import (
	"time"
	"strconv"
)

// A version string from discovery
// See https://github.com/geteduroam/cattenbak/blob/481e243f22b40e1d8d48ecac2b85705b8cb48494/cattenbak.py#L116
type Seq struct {
	Timestamp time.Time 
	Offset int
}

func NewSeq(seq int) (*Seq, error) {
	raw := strconv.Itoa(seq)
	i := len(raw) - 2
	t, o := raw[:i], raw[i:]
	// Defined as year, month, day
	p, err := time.Parse("20060102", t)
	if err != nil {
		return nil, err
	}
	// Parse offset
	offset, err := strconv.Atoi(o)
	if err != nil {
		return nil, err
	}
	return &Seq{Timestamp: p, Offset: offset}, nil
}

func (s Seq) After(o Seq) bool {
	// Timestamp is already higher
	// This means day(s) after so it must be newer
	if s.Timestamp.After(o.Timestamp) {
		return true
	}
	// Timestamp is lower
	// This means day(s) before so it must be older
	if s.Timestamp.Before(o.Timestamp) {
		return false
	}
	// Days are the same, the offset should be higher
	return s.Offset > o.Offset
}
