package flagrun

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
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
			o := &testRunner{}
			options := []FlagrunOptions{}
			if tt.wantArgs {
				options = append(options, ArgsRequired())
			}
			if tt.version != "" {
				options = append(options, Version(tt.version))
			}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			msg, code := internalGo(tt.args, &stdout, &stderr, o, options...)
			stdoutStr := stdout.String()
			stderrStr := stderr.String()

			assert.Equal(t, tt.wantMsg, msg, "%s msg", tt.name)
			assert.Contains(t, stdoutStr, tt.wantStdout, "%s stdout", tt.name)
			assert.Contains(t, stderrStr, tt.wantStderr, "%s stderr", tt.name)
			assert.Equal(t, code, tt.wantCode, "%s code", tt.name)
			if tt.wantArgs {
				assert.Equal(t, tt.wantArgv, o.argv, "%s argv", tt.name)
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
	msg, code := internalGo([]string{}, &stdout, &stderr, o)
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	assert.Equal(t, "", msg)
	assert.Contains(t, stdoutStr, "")
	assert.Contains(t, stderrStr, "the required flag `-r, --required")
	assert.Equal(t, UNKNOWN, code)
}
