package flagrun

import (
	"bytes"
	"os"
	"testing"

	"github.com/mackerelio/checkers"
	"github.com/stretchr/testify/assert"
)

type testChecker struct {
	Version bool `short:"v" long:"version" description:"Show version"`
	status  checkers.Status
	msg     string
}

func (c *testChecker) Run(_ []string) *checkers.Checker {
	return checkers.NewChecker(c.status, c.msg)
}

func TestInternalChecker(t *testing.T) {
	tests := []struct {
		name       string
		status     checkers.Status
		msg        string
		wantStdout string
		wantCode   int
	}{
		{
			name:       "ok",
			status:     checkers.OK,
			msg:        "service is ok",
			wantStdout: " OK: service is ok",
			wantCode:   OK,
		},
		{
			name:       "warning",
			status:     checkers.WARNING,
			msg:        "service is warning",
			wantStdout: " WARNING: service is warning",
			wantCode:   WARNING,
		},
		{
			name:       "critical",
			status:     checkers.CRITICAL,
			msg:        "service is critical",
			wantStdout: " CRITICAL: service is critical",
			wantCode:   CRITICAL,
		},
		{
			name:       "unknown",
			status:     checkers.UNKNOWN,
			msg:        "service is unknown",
			wantStdout: " UNKNOWN: service is unknown",
			wantCode:   UNKNOWN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := buildFlagrun()
			o := &testChecker{status: tt.status, msg: tt.msg}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			msg, code := f.internalChecker([]string{}, &stdout, &stderr, o)
			stdoutStr := stdout.String()

			assert.Equal(t, "", stdoutStr)
			assert.Equal(t, "flagrun.test"+tt.wantStdout, msg)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestInternalCheckerHelp(t *testing.T) {
	f := buildFlagrun()
	o := &testChecker{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	msg, code := f.internalChecker([]string{"--help"}, &stdout, &stderr, o)
	stdoutStr := stdout.String()

	assert.Equal(t, "", msg)
	assert.Equal(t, OK, code)
	assert.Contains(t, stdoutStr, "Show version")
	assert.Contains(t, stdoutStr, "Show this help message")
}

func TestInternalCheckerVersion(t *testing.T) {
	f := buildFlagrun(Version("1.0.0"))
	o := &testChecker{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	msg, code := f.internalChecker([]string{"--version"}, &stdout, &stderr, o)
	stdoutStr := stdout.String()

	assert.Equal(t, "", msg)
	assert.Equal(t, OK, code)
	assert.Contains(t, stdoutStr, "test-1.0.0")
}

func TestCheck(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"test"}

	o := &testChecker{status: checkers.OK, msg: "ok"}
	code := Check(o)

	assert.Equal(t, OK, code)
}
