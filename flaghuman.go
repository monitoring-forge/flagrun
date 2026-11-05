package flagrun

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

var ErrBytesOverflow = fmt.Errorf("Bytes value overflows int64")

// Bytes is a custom type to handle human-readable byte inputs
type Bytes uint64

// UnmarshalFlag parses human inputs like "10MB" or "2GiB" using go-humanize
func (hb *Bytes) UnmarshalFlag(value string) error {
	bytes, err := humanize.ParseBytes(value)
	if err != nil {
		return err
	}
	*hb = Bytes(bytes)
	return nil
}

func (hb Bytes) Int64() (int64, error) {
	if hb > Bytes(^uint64(0)>>1) {
		return 0, ErrBytesOverflow
	}
	return int64(hb), nil
}
