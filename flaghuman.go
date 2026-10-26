package flagrun

import "github.com/dustin/go-humanize"

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
