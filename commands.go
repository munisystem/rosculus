package main

import (
	"github.com/mitchellh/cli"
	"github.com/munisystem/rstack/command"
)

func Commands(meta *command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"new": func() (cli.Command, error) {
			return &command.NewCommand{
				Meta: *meta,
			}, nil
		},
		"rotate": func() (cli.Command, error) {
			return &command.RotateCommand{
				Meta: *meta,
			}, nil
		},
		"rollback": func() (cli.Command, error) {
			return &command.RollbackCommand{
				Meta: *meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:     *meta,
				Version:  Version,
				Revision: GitCommit,
				Name:     Name,
			}, nil
		},
	}
}
