package command

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/munisystem/rstack/aws/rds"
	"github.com/munisystem/rstack/deployment"
)

type RotateCommand struct {
	Meta
}

func (c *RotateCommand) Run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, errors.New("too few arguments"))
		return 1
	} else if len(args) > 1 {
		fmt.Fprintln(os.Stderr, errors.New("too many arguments"))
		return 1
	}

	name := args[0]

	bucket := os.Getenv("AWS_S3_BUCKET_NAME")
	if bucket == "" {
		fmt.Fprintln(os.Stderr, errors.New("Please set s3 bucket name in AWS_S3_BUCKET_NAME"))
		return 1
	}

	dep, err := deployment.Load(bucket, name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Printf("Launch new RDS Instance '%s' from '%s'\n", dep.Previous.InstanceIdentifier, dep.SourceDBInstanceIdentifier)
	dbInstance, err := rds.CopyInstance(
		dep.SourceDBInstanceIdentifier,
		dep.Previous.InstanceIdentifier,
		dep.DBInstanceClass,
		dep.PubliclyAccessible,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println("Launched. Please Wait RDS Instance ready")

	for {
		if *dbInstance.DBInstanceStatus == "available" {
			fmt.Printf("%s is ready\n", *dbInstance.DBName)
			break
		}

		fmt.Print(".")
		time.Sleep(30 * time.Second)
	}

	prev := deployment.Previous{
		InstanceIdentifier: dep.Current.InstanceIdentifier,
		Endpoint:           dep.Current.Endpoint,
	}

	cur := deployment.Current{
		InstanceIdentifier: dep.Previous.InstanceIdentifier,
		Endpoint:           *dbInstance.Endpoint.Address,
	}

	dep.Current = cur
	dep.Previous = prev
	dep.New(bucket, name)

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
