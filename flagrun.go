package flagrun

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"

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
	msg, code := internalGo(os.Args[1:], os.Stdout, os.Stderr, opt, options...)
	if msg != "" {
		if code == OK {
			_ = println(os.Stdout, msg)
		} else {
			_ = println(os.Stderr, msg)
		}
	}
	return code
}

// hasBooleanVersionField checks if the struct has a Version field of type bool and its value is true
func hasBooleanVersionField(opt Runner) bool {
	t := reflect.TypeOf(opt)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	field, ok := t.FieldByName("Version")
	if !ok {
		return false
	}
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return ok && field.Type.Kind() == reflect.Bool && v.FieldByName("Version").Bool()
}

func internalGo(
	argv []string,
	stdout io.Writer,
	stderr io.Writer,
	opt Runner,
	options ...FlagrunOptions,
) (string, int) {
	f := &Flagrun{
		Commit:  "dev",
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
		return "", OK
	} else if flags.WroteHelp(err) {
		fmt.Fprintf(stdout, "%v\n", err)
		return "", OK
	} else if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return "", UNKNOWN
	} else if f.ArgsRequired && len(args) == 0 {
		fmt.Fprintf(stderr, "command is required\n")
		psr.WriteHelp(stderr)
		return "", UNKNOWN
	}

	return opt.Run(args)
}
