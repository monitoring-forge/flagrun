package flagrun

import "regexp"

type Re struct {
	value  string
	regexp *regexp.Regexp
}

func (fr *Re) UnmarshalFlag(value string) error {
	if value == "" {
		fr.value = ""
		fr.regexp = nil
		return nil
	}
	re, err := regexp.Compile(value)
	if err != nil {
		return err
	}
	fr.value = value
	fr.regexp = re
	return nil
}

func (fr *Re) String() string {
	return fr.value
}

func (fr *Re) Regexp() *regexp.Regexp {
	return fr.regexp
}

func (fr *Re) Present() bool {
	return fr.regexp != nil
}
