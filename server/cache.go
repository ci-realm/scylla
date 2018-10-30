package server

import (
	"bytes"
	"fmt"
	"os/exec"
)

type hasResultNixPaths interface {
	resultNixPaths() []string
	runCmd(*exec.Cmd) (*bytes.Buffer, error)
}

// NOTE: this has to SSH to the worker, it'll push the whole store!
// nixCopyURL example: "s3://scylla-cache?region=eu-central-1"
func copyResultsToNixStore(j hasResultNixPaths, nixCopyURL string) error {
	if nixCopyURL == "" {
		return nil
	}

	for _, nixStorePath := range j.resultNixPaths() {
		_, err := j.runCmd(exec.Command(
			"ssh", "root@3.120.166.103",
			"nix", "copy", nixStorePath, "--to", nixCopyURL,
		))
		if err != nil {
			return fmt.Errorf("Copying %s to %s: %s", nixStorePath, nixCopyURL, err)
		}
	}

	return nil
}

// NOTE: intermediate results won't be cached unless we run with watch mode
func copyResultsToCachix(j hasResultNixPaths, cacheName string) error {
	if cacheName == "" {
		return nil
	}

	for _, nixStorePath := range j.resultNixPaths() {
		command := exec.Command("cachix", "push", cacheName)
		command.Stdin = bytes.NewReader([]byte(nixStorePath))
		_, err := j.runCmd(command)
		if err != nil {
			return fmt.Errorf("Copying %s to %s: %s", nixStorePath, cacheName, err)
		}
	}

	return nil
}
