package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/munisystem/rstack/deployment"
)

type NewCommand struct {
	Meta

	name                       string
	sourceDBInstanceIdentifier string
	dbInstanceIdentifierBase   string
	publiclyAccessible         bool
	dbInstanceClass            string
	securityGroupsString       string
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

	var securityGroups []string
	if c.securityGroupsString != "" {
		securityGroups = strings.Split(c.securityGroupsString, ",")
	}

	dep := &deployment.Deployment{
		SourceDBInstanceIdentifier: c.sourceDBInstanceIdentifier,
		PubliclyAccessible:         c.publiclyAccessible,
		DBInstanceClass:            c.dbInstanceClass,
		SecurityGroups:             securityGroups,
		Current: deployment.Current{
			InstanceIdentifier: c.dbInstanceIdentifierBase + "-blue",
			Endpoint:           "",
		},
		Previous: deployment.Previous{
			InstanceIdentifier: c.dbInstanceIdentifierBase + "-green",
			Endpoint:           "",
		},
	}

	if err := dep.Put(bucket, c.name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func (c *NewCommand) parseArgs(args []string) error {
	flag := flag.NewFlagSet("rstack", flag.ContinueOnError)

	flag.StringVar(&c.sourceDBInstanceIdentifier, "source-db-instance-identifier", "", "SourceDBInstanceIdentifier")
	flag.StringVar(&c.dbInstanceIdentifierBase, "db-instance-identifier-base", "", "DBInstanceIdentifierBase")
	flag.BoolVar(&c.publiclyAccessible, "publicly-accessible", true, "PubliclyAccessible")
	flag.StringVar(&c.dbInstanceClass, "db-instance-class", "db.m3.medium", "DBInstanceClass")
	flag.StringVar(&c.securityGroupsString, "security-groups", "", "SecurityGroups")

	if err := flag.Parse(args); err != nil {
		return err
	}

	if c.sourceDBInstanceIdentifier == "" {
		return errors.New("Please specify original DB instance identifier")
	}

	if c.dbInstanceIdentifierBase == "" {
		return errors.New("Please specify DB instance identifier base")
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
