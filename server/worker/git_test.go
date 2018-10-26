package worker

import (
	"testing"
)

func TestGit(t *testing.T) {
	worker := Worker{}

	args := Args{
		Location:    "/home/manveru/go/src/github.com/manveru/scylla",
		RawSchedule: "* * * * *",
		ShellNix: `
with import <nixpkgs> {};
mkShell {
  buildInputs = [ tree ];
}
`,
	}

	worker.Run(args)
}
