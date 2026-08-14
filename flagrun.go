package flagrun

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/jessevdk/go-flags"
)

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

type Runner interface {
	// Run executes the command with the provided flags and arguments.
	Run([]string) (string, int)
}

type Flagrun struct {
	ArgsRequired bool
	Version      string
	Commit       string
	AlwaysStdout bool
}

type FlagrunOptions func(*Flagrun)

func Version(version string) FlagrunOptions {
	return func(f *Flagrun) {
		if version != "" {
			f.Version = version
		}
	}
}

func Commit(commit string) FlagrunOptions {
	return func(f *Flagrun) {
		if commit != "" {
			f.Commit = commit
		}
	}
}

func ArgsRequired() FlagrunOptions {
	return func(f *Flagrun) {
		f.ArgsRequired = true
	}
}

func AlwaysStdout() FlagrunOptions {
	return func(f *Flagrun) {
		f.AlwaysStdout = true
	}
}

func printLine(w io.Writer, s string) error {
	if w == nil {
		return nil
	}
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	_, err := io.WriteString(w, s)
	return err
}

func Go(opt Runner, options ...FlagrunOptions) int {
	f, msg, code := internalGo(os.Args[1:], os.Stdout, os.Stderr, opt, options...)
	if msg != "" {
		if code == OK || f.AlwaysStdout {
			_ = printLine(os.Stdout, msg)
		} else {
			_ = printLine(os.Stderr, msg)
		}
	}
	return code
}

// hasBooleanVersionField checks if the struct has a Version field of type bool and its value is true
func hasBooleanVersionField(opt Runner) bool {
	if opt == nil {
		return false
	}
	t := reflect.TypeOf(opt)
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Pointer {
		if t.Elem().Kind() != reflect.Struct {
			return false
		}
		t = t.Elem()
	}
	field, ok := t.FieldByName("Version")
	if !ok {
		return false
	}
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Pointer {
		if v.Elem().Kind() != reflect.Struct {
			return false
		}
		v = v.Elem()
	}
	return ok && field.Type.Kind() == reflect.Bool && v.FieldByName("Version").Bool()
}

func buildCommitHash() string {
	commit := "dev"
	dirty := false
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				commit = setting.Value
			} else if setting.Key == "vcs.modified" && setting.Value == "true" {
				dirty = true
			}
		}
	}
	if len(commit) > 7 {
		commit = commit[:7]
	}
	if dirty {
		commit += "-dirty"
	}
	return commit
}

func internalGo(
	argv []string,
	stdout io.Writer,
	stderr io.Writer,
	opt Runner,
	options ...FlagrunOptions,
) (*Flagrun, string, int) {
	f := &Flagrun{
		Commit:  buildCommitHash(),
		Version: "unknown",
	}
	for _, option := range options {
		option(f)
	}

	psr := flags.NewParser(opt, flags.HelpFlag|flags.PassDoubleDash)
	if f.ArgsRequired {
		psr.Usage = "[OPTIONS] -- command [args...]"
	}
	args, err := psr.ParseArgs(argv)
	// opt has a Version field, print version and exit
	if hasBooleanVersionField(opt) {
		fmt.Fprintf(
			stdout,
			"%s-%s\n%s/%s, %s, %s\n",
			psr.Name,
			f.Version,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version(),
			f.Commit)
		return f, "", OK
	} else if flags.WroteHelp(err) {
		fmt.Fprintf(stdout, "%v\n", err)
		return f, "", OK
	} else if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return f, "", UNKNOWN
	} else if f.ArgsRequired && len(args) == 0 {
		fmt.Fprintf(stderr, "command is required\n")
		psr.WriteHelp(stderr)
		return f, "", UNKNOWN
	}
	msg, code := opt.Run(args)
	return f, msg, code
}
