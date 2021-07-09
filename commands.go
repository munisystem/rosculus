package main

import (
	"github.com/mitchellh/cli"
	"github.com/munisystem/rosculus/command"
)

func Commands(meta *command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"rotate": func() (cli.Command, error) {
			return &command.RotateCommand{
				Meta: *meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:     *meta,
				Version:  Version,
				Revision: Revision,
				Name:     Name,
			}, nil
		},
	}
}
