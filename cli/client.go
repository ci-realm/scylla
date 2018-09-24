package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
)

type Evaluation struct {
	Source []byte
}

func main() {
	args := os.Args()
	app := cli.NewApp()
	app.Name = "scy"
	app.Commands = []cli.Commad{
		{
			Name:        "push",
			Usage:       "push rev",
			Description: "Push the given revision to the CI server an initiate a build",
			Flags:       []Flag{},
			Action: func(c *cli.Context) error {
				fmt.Println("pushing", c.String("rev"))
			},
		},
	}

	exec.Command("git", "archive", "master")
}
