package flagrun

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRunner struct {
	Version bool `short:"v" long:"version" description:"Show version"`
	argv    []string
}

func (r *testRunner) Run(argv []string) (string, int) {
	r.argv = argv
	return "Test runner executed", OK
}

func TestInternalGo(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		version    string
		wantMsg    string
		wantStdout string
		wantStderr string
		wantCode   int
		wantArgs   bool
		wantArgv   []string
	}{
		{
			name:       "no arguments",
			args:       []string{},
			wantMsg:    "",
			wantStderr: "command is required\n",
			wantCode:   UNKNOWN,
			wantArgs:   true,
		},
		{
			name:       "help flag",
			args:       []string{"--help"},
			wantMsg:    "",
			wantStdout: "Usage:\n  flagrun.test [OPTIONS]\n\nApplication Options:\n  -v, --version  Show version\n\nHelp Options:\n  -h, --help     Show this help message\n",
			wantCode:   OK,
		},
		{
			name:       "version flag",
			args:       []string{"--version"},
			version:    "1.0.0",
			wantMsg:    "",
			wantStdout: "flagrun.test-1.0.0",
			wantCode:   OK,
		},
		{
			name:       "valid arguments",
			args:       []string{"-k"},
			wantMsg:    "",
			wantStdout: "",
			wantStderr: "unknown flag `k'",
			wantCode:   UNKNOWN,
		},
		{
			name:       "valid arguments with args required",
			args:       []string{"arg1", "arg2"},
			wantMsg:    "Test runner executed",
			wantStdout: "",
			wantStderr: "",
			wantCode:   OK,
			wantArgs:   true,
			wantArgv:   []string{"arg1", "arg2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &testRunner{}
			options := []FlagrunOptions{}
			if tt.wantArgs {
				options = append(options, ArgsRequired())
			}
			if tt.version != "" {
				options = append(options, Version(tt.version))
			}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			f := buildFlagrun(options...)
			msg, code := internalGo(f, tt.args, &stdout, &stderr, runner)
			stdoutStr := stdout.String()
			stderrStr := stderr.String()
			require.NotNil(t, f, "%s Flagrun instance should not be nil", tt.name)
			assert.Equal(t, tt.wantMsg, msg, "%s msg", tt.name)
			assert.Contains(t, stdoutStr, tt.wantStdout, "%s stdout", tt.name)
			assert.Contains(t, stderrStr, tt.wantStderr, "%s stderr", tt.name)
			assert.Equal(t, code, tt.wantCode, "%s code", tt.name)
			if tt.wantArgs {
				assert.Equal(t, tt.wantArgv, runner.argv, "%s argv", tt.name)
			}
		})
	}
}

type requiredRunner struct {
	Version  bool   `short:"v" long:"version" description:"Show version"`
	Required string `short:"r" long:"required" description:"Required parameter" required:"true"`
}

func (r *requiredRunner) Run(_ []string) (string, int) {
	return "Test runner executed", OK
}

func TestInternalGoWithRequiredParameters(t *testing.T) {
	o := &requiredRunner{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	f := buildFlagrun()
	msg, code := internalGo(f, []string{}, &stdout, &stderr, o)
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	assert.NotNil(t, f, "Flagrun instance should not be nil")
	assert.Equal(t, "", fmt.Sprintf("%v", msg), "msg should be empty")
	assert.Contains(t, stdoutStr, "")
	assert.Contains(t, stderrStr, "the required flag `-r, --required")
	assert.Equal(t, UNKNOWN, code)
}

type versionTrueRunner struct {
	Version bool `short:"v" long:"version" description:"Show version"`
}

type versionFalseRunner struct {
	Version bool `short:"v" long:"version" description:"Show version"`
}

type noVersionRunner struct{}

type stringVersionRunner struct {
	Version string `short:"v" long:"version" description:"Show version"`
}

func (r *versionTrueRunner) Run(_ []string) (string, int)  { return "", OK }
func (r *versionFalseRunner) Run(_ []string) (string, int) { return "", OK }
func (r *noVersionRunner) Run(_ []string) (string, int)    { return "", OK }
func (r *stringVersionRunner) Run(_ []string) (string, int) {
	return "", OK
}

func TestHasBooleanVersionField(t *testing.T) {
	tests := []struct {
		name string
		opt  Runner[string]
		want bool
	}{
		{
			name: "Version field is true",
			opt:  &versionTrueRunner{Version: true},
			want: true,
		},
		{
			name: "Version field is false",
			opt:  &versionFalseRunner{Version: false},
			want: false,
		},
		{
			name: "no Version field",
			opt:  &noVersionRunner{},
			want: false,
		},
		{
			name: "Version field is string",
			opt:  &stringVersionRunner{Version: "1.0.0"},
			want: false,
		},
		{
			name: "nil runner",
			opt:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasBooleanVersionField(tt.opt))
		})
	}
}

type anyMessageRunner struct {
	Switch bool `short:"s" long:"switch" description:"A boolean switch"`
}

func (r *anyMessageRunner) Run(_ []string) (any, int) {
	if r.Switch {
		return fmt.Errorf("Switch is %v", r.Switch), CRITICAL
	}
	return "Switch is OFF", OK
}

func TestInternalGoWithAnyMessageType(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantMsg  string
		wantCode int
	}{
		{
			name:     "Switch is OFF",
			args:     []string{},
			wantMsg:  "Switch is OFF",
			wantCode: OK,
		},
		{
			name:     "Switch is ON",
			args:     []string{"--switch"},
			wantMsg:  "Switch is true",
			wantCode: CRITICAL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &anyMessageRunner{}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			f := buildFlagrun()
			msg, code := internalGo(f, tt.args, &stdout, &stderr, o) // T is inferred
			assert.Equal(t, tt.wantMsg, msg, "%s msg", tt.name)
			assert.Equal(t, code, tt.wantCode, "%s code", tt.name)
		})
	}
}
