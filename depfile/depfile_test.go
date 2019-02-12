package depfile

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

var parseTests = []struct {
	testName    string
	data        string
	expectError string
	expectFile  *File
}{{
	testName: "success",
	data: `
build std
test example.com/xx/...
build foo.com/bar/...
test example.com/foo
`,
	expectFile: &File{
		Build: []string{
			"foo.com/bar/...",
			"std",
		},
		Test: []string{
			"example.com/foo",
			"example.com/xx/...",
		},
	},
}, {
	testName: "bad verb",
	data: `
xxx std
`,
	expectError: `unknown verb "xxx"`,
}, {
	testName: "wrong field count",
	data: `
test foo bar
`,
	expectError: `wrong number of fields on line "test foo bar"`,
}}

func TestParse(t *testing.T) {
	c := qt.New(t)
	for _, test := range parseTests {
		c.Run(test.testName, func(c *qt.C) {
			f, err := Parse([]byte(test.data))
			if test.expectError != "" {
				c.Assert(err, qt.ErrorMatches, test.expectError)
				return
			}
			c.Assert(err, qt.Equals, nil)
			c.Assert(f, qt.DeepEquals, test.expectFile)
		})
	}
}

var formatTests = []struct {
	testName string
	f        *File
	expect   string
}{{
	testName: "empty",
	f:        new(File),
	expect:   "",
}, {
	testName: "build only",
	f: &File{
		Build: []string{"foo", "bar"},
	},
	expect: "build foo\nbuild bar\n",
}, {
	testName: "test only",
	f: &File{
		Test: []string{"foo", "bar"},
	},
	expect: "test foo\ntest bar\n",
}, {
	testName: "build and test",
	f: &File{
		Build: []string{"arble", "bletch"},
		Test:  []string{"foo", "bar"},
	},
	expect: "build arble\nbuild bletch\n\ntest foo\ntest bar\n",
}}

func TestFormat(t *testing.T) {
	c := qt.New(t)
	for _, test := range formatTests {
		c.Run(test.testName, func(c *qt.C) {
			data := Format(test.f)
			c.Assert(string(data), qt.Equals, test.expect)
		})
	}
}
