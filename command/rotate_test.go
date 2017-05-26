package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestRotateCommand_implement(t *testing.T) {
	var _ cli.Command = &RotateCommand{}
}
