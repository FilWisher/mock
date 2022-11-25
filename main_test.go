package main

import (
	"testing"
	"strings"
	"os"
	"path/filepath"
	"flag"
)

var regenerate = flag.Bool("regenerate", false, "regenerate the golden test files")

func noerr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerate(t *testing.T) {

	tests, err := filepath.Glob("fixtures/*.in")
	noerr(t, err)

	for _, test := range tests {
		base := strings.TrimSuffix(filepath.Base(test), filepath.Ext(test))
		t.Run(base, func(t *testing.T) {
			in, err := os.ReadFile("fixtures/" + base + ".in")
			noerr(t, err)

			var b strings.Builder
			err = generate(string(in), &b)
			noerr(t, err)

			got := b.String()

			// If the tests are run with the -regenerate flag,
			// write the generated output to the .out file.
			if *regenerate {
				err := os.WriteFile("fixtures/" + base  + ".out", []byte(got), 0644)
				noerr(t, err)
				return
			}

			expected, err := os.ReadFile("fixtures/" + base + ".out")
			noerr(t, err)

			if got != string(expected) {
				t.Errorf("output does not match expected. expected:\n%s\ngot:\n%s\n", expected, got)
			}
		})
	}
}
