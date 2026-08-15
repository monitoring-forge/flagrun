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
	"github.com/mackerelio/checkers"
)

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

type Runner[T any] interface {
	// Run executes the command with the provided flags and arguments.
	Run([]string) (T, int)
}

type Checker interface {
	// Check executes the command with the provided flags and arguments.
	Run([]string) *checkers.Checker
}

type Shipper interface {
	// Run executes the command with the provided flags and arguments.
	Run([]string)
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

// hasBooleanVersionField checks if the struct has a Version field of type bool and its value is true
func hasBooleanVersionField(opt any) bool {
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

func buildFlagrun(options ...FlagrunOptions) *Flagrun {
	f := &Flagrun{
		Commit:  buildCommitHash(),
		Version: "unknown",
	}
	for _, option := range options {
		option(f)
	}
	return f
}

func nullint(i int) *int {
	return &i
}

func (f *Flagrun) parseArgs(argv []string, stdout, stderr io.Writer, opt any) ([]string, *int) {
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
		return nil, nullint(OK)
	} else if flags.WroteHelp(err) {
		fmt.Fprintf(stdout, "%v\n", err)
		return nil, nullint(OK)
	} else if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return nil, nullint(UNKNOWN)
	} else if f.ArgsRequired && len(args) == 0 {
		fmt.Fprintf(stderr, "command is required\n")
		psr.WriteHelp(stderr)
		return nil, nullint(UNKNOWN)
	}
	return args, nil
}

func internalGo[T any](
	f *Flagrun,
	argv []string,
	stdout io.Writer,
	stderr io.Writer,
	opt Runner[T],
) (string, int) {
	args, c := f.parseArgs(argv, stdout, stderr, opt)
	if c != nil {
		return "", *c
	}
	msg, code := opt.Run(args)
	return fmt.Sprintf("%v", msg), code
}

// Checker return *checkers.Checker
func (f *Flagrun) internalChecker(
	argv []string,
	stdout io.Writer,
	stderr io.Writer,
	opt Checker,
) (string, int) {
	args, c := f.parseArgs(argv, stdout, stderr, opt)
	if c != nil {
		return "", *c
	}
	f.AlwaysStdout = true
	chk := opt.Run(args)
	if chk == nil {
		return "UNKNOWN: checker returned nil", UNKNOWN
	}
	switch chk.Status {
	case checkers.OK:
	    return chk.String(), OK
	case checkers.WARNING:
	    return chk.String(), WARNING
	case checkers.CRITICAL:
	    return chk.String(), CRITICAL
	default:
	    return chk.String(), UNKNOWN
	}
}

func (f *Flagrun) internalShipper(
	argv []string,
	stdout io.Writer,
	stderr io.Writer,
	opt Shipper,
) int {
	args, c := f.parseArgs(argv, stdout, stderr, opt)
	if c != nil {
		return *c
	}
	opt.Run(args)
	return OK
}

func Go[T any](opt Runner[T], options ...FlagrunOptions) int {
	f := buildFlagrun(options...)
	msg, code := internalGo(f, os.Args[1:], os.Stdout, os.Stderr, opt)
	if msg != "" {
		if code == OK || f.AlwaysStdout {
			_ = printLine(os.Stdout, msg)
		} else {
			_ = printLine(os.Stderr, msg)
		}
	}
	return code
}

func Check(opt Checker, options ...FlagrunOptions) int {
	f := buildFlagrun(options...)
	msg, code := f.internalChecker(os.Args[1:], os.Stdout, os.Stderr, opt)
	if msg != "" {
		_ = printLine(os.Stdout, msg)
	}
	return code
}

func Ship(opt Shipper, options ...FlagrunOptions) int {
	f := buildFlagrun(options...)
	return f.internalShipper(os.Args[1:], os.Stdout, os.Stderr, opt)
}
