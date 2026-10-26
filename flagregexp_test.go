package flagrun

import "testing"

func TestFlagRegexp(t *testing.T) {
	testCases := []struct {
		input    string
		err      bool
		expected bool
	}{
		{"", false, false},
		{"^example", false, true},
		{"invalid[", true, false},
	}
	for _, tc := range testCases {
		fr := &Re{}
		err := fr.UnmarshalFlag(tc.input)
		if !tc.err && err != nil {
			t.Errorf("unexpected error for input %q: %v", tc.input, err)
		}
		if tc.err && err == nil {
			t.Errorf("expected error for input %q, but got none", tc.input)
		}
		if fr.Present() != tc.expected {
			t.Errorf("input: %q, expected Present: %v, got: %v", tc.input, tc.expected, fr.Present())
		}
	}
}
