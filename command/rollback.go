package command

import (
	"strings"
)

type RollbackCommand struct {
	Meta
}

func (c *RollbackCommand) Run(args []string) int {
	// Write your code here

	return 0
}

func (c *RollbackCommand) Synopsis() string {
	return ""
}

func (c *RollbackCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
