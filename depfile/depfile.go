// Package depfile implements support for reading and
// writing go.dep files.
// NOTE this currently supports only a very limited subset
// of the go.dep file format described in the README.
package depfile

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/errgo.v2/fmt/errors"
)

// File represents a go.dep file.
type File struct {
	Build []string
	Test  []string
}

// Parse parses the data, reported in errors as being from file,
// into a File struct.
//
// Currently this supports only packages on single lines, each
// prefixed by either "build" or "test".
//
// TODO implement the full file format with comments.
func Parse(data []byte) (*File, error) {
	f := new(File)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, errors.Newf("wrong number of fields on line %q", line)
		}
		switch fields[0] {
		case "build":
			f.Build = append(f.Build, fields[1])
		case "test":
			f.Test = append(f.Test, fields[1])
		default:
			return nil, errors.Newf("unknown verb %q", fields[0])
		}
	}
	sort.Strings(f.Build)
	sort.Strings(f.Test)
	return f, nil
}

// Format returns the contents of the given file.
func Format(f *File) []byte {
	var buf bytes.Buffer
	for _, p := range f.Build {
		fmt.Fprintf(&buf, "build %s\n", p)
	}
	if len(f.Build) > 0 && len(f.Test) > 0 {
		buf.WriteByte('\n')
	}
	for _, p := range f.Test {
		fmt.Fprintf(&buf, "test %s\n", p)
	}
	return buf.Bytes()
}
