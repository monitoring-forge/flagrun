package flagrun

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testShipper struct {
	Version bool `short:"v" long:"version" description:"Show version"`
	ran     bool
}

func (s *testShipper) Run(_ []string) {
	s.ran = true
}

func TestInternalShipper(t *testing.T) {
	f := buildFlagrun()
	o := &testShipper{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	f.internalShipper([]string{}, &stdout, &stderr, o)

	assert.True(t, o.ran)
}

func TestInternalShipperHelp(t *testing.T) {
	f := buildFlagrun()
	o := &testShipper{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	f.internalShipper([]string{"--help"}, &stdout, &stderr, o)
	stdoutStr := stdout.String()

	assert.False(t, o.ran)
	assert.Contains(t, stdoutStr, "Show version")
	assert.Contains(t, stdoutStr, "Show this help message")
}

func TestInternalShipperVersion(t *testing.T) {
	f := buildFlagrun(Version("1.0.0"))
	o := &testShipper{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	f.internalShipper([]string{"--version"}, &stdout, &stderr, o)
	stdoutStr := stdout.String()

	assert.False(t, o.ran)
	assert.Contains(t, stdoutStr, "test-1.0.0")
}

func TestShip(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"test"}

	o := &testShipper{}
	code := Ship(o)

	assert.Equal(t, OK, code)
	assert.True(t, o.ran)
}
