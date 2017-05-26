package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestRollbackCommand_implement(t *testing.T) {
	var _ cli.Command = &RollbackCommand{}
}
