package flagrun

import "testing"

func TestHumanBytesUnmarshalFlag(t *testing.T) {
	var hb Bytes
	if err := hb.UnmarshalFlag("10MB"); err != nil {
		t.Fatalf("UnmarshalFlag error: %v", err)
	}
	if hb != 10*1000*1000 {
		t.Errorf("Bytes = %d; want 10000000", hb)
	}
}

func TestHumanBytesUnmarshalFlagInvalid(t *testing.T) {
	var hb Bytes
	if err := hb.UnmarshalFlag("invalid"); err == nil {
		t.Error("UnmarshalFlag should return error for invalid input")
	}
}

func TestHumanBytesInt64(t *testing.T) {
	var hb Bytes = 10 * 1000 * 1000
	val, err := hb.Int64()
	if err != nil {
		t.Fatalf("Int64 error: %v", err)
	}
	if val != 10*1000*1000 {
		t.Errorf("Int64 = %d; want 10000000", val)
	}
}

func TestHumanBytesInt64Overflow(t *testing.T) {
	var hb Bytes = ^Bytes(0) // Max uint64
	_, err := hb.Int64()
	if err != ErrBytesOverflow {
		t.Errorf("Int64 error = %v; want ErrBytesOverflow", err)
	}
}
