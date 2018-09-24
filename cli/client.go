package cli

import (
	"fmt"
	"os/exec"

	"github.com/codegangsta/cli"
)

type Evaluation struct {
	Source []byte
}

func main() {
	app := cli.NewApp()
	app.Name = "scy"
	app.Commands = []cli.Command{
		{
			Name:        "push",
			Usage:       "push rev",
			Description: "Push the given revision to the CI server an initiate a build",
			Flags:       []cli.Flag{},
			Action: func(c *cli.Context) error {
				fmt.Println("pushing", c.String("rev"))
				return nil
			},
		},
	}

	exec.Command("git", "archive", "master")
}
