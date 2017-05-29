package command

import (
	"strings"
	"flag"
	"errors"
	"fmt"
	"os"
	"github.com/munisystem/rstack/deployment"
)

type NewCommand struct {
	Meta

	name string
	sourceDBIdentifier string
	publiclyAccessible bool
	dbInstanceClass string
}

func (c *NewCommand) Run(args []string) int {
	if err := c.parseArgs(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	bucket := os.Getenv("AWS_S3_BUCKET_NAME")
	if bucket == "" {
		fmt.Fprintln(os.Stderr, errors.New("Please set s3 bucket name in AWS_S3_BUCKET_NAME"))
		return 1
	}

	dep := &deployment.Deployment{
		SourceDBInstanceIdentifier: c.sourceDBIdentifier,
		PubliclyAccessible: c.publiclyAccessible,
		DBInstanceClass: c.dbInstanceClass,
		Current: deployment.Current {
			InstanceIdentifier: "",
			Endpoint: "",
		},
		Previous: deployment.Previous {
			InstanceIdentifier: "",
			Endpoint: "",
		},
	}

	if err := dep.New(bucket, c.name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func (c *NewCommand) parseArgs(args []string) error {
	flag := flag.NewFlagSet("rstack", flag.ContinueOnError)

	flag.StringVar(&c.sourceDBIdentifier, "source-db-instance-identifier", "", "SourceDBInstanceIdentifier")
	flag.BoolVar(&c.publiclyAccessible, "publicly-accessible", true, "PubliclyAccessible")
	flag.StringVar(&c.dbInstanceClass, "db-instance-class", "db.m3.medium", "DBInstanceClass")

	if err := flag.Parse(args); err != nil {
		return err
	}

	if c.sourceDBIdentifier == "" {
		return errors.New("Please specify original DB instance identifier")
	}

	if 0 < flag.NArg() {
		c.name = flag.Arg(0)
	}

	if c.name == "" {
		return errors.New("Please specify deployment name")
	}

	return nil
}

func (c *NewCommand) Synopsis() string {
	return ""
}

func (c *NewCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
