package command

import (
	"strings"
)

type RotateCommand struct {
	Meta
}

func (c *RotateCommand) Run(args []string) int {
	// Write your code here

	return 0
}

func (c *RotateCommand) Synopsis() string {
	return ""
}

func (c *RotateCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
