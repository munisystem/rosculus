package command

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/munisystem/rstack/aws/rds"
	"github.com/munisystem/rstack/deployment"
	"github.com/munisystem/rstack/dnsimple"
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
		return 1
	}
	fmt.Println("Launched. Please Wait RDS Instance ready")

	errCh := make(chan error, 2)
	go func() {
		errCh <- rds.WaitReady(*dbInstance.DBInstanceIdentifier)
	}()

	go func() {
	loop:
		for {
			if len(errCh) > 0 {
				fmt.Print("\n")
				break loop
			}

			fmt.Print(".")
			time.Sleep(30 * time.Second)
		}

	}()

	err = <-errCh
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("%s is ready\n", *dbInstance.DBName)

	if err = rds.AddSecurityGroups(*dbInstance.DBInstanceIdentifier, dep.SecurityGroups); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Println("Attached security groups %s", dep.SecurityGroups)

	dbInstances, err := rds.DescribeDBInstances(*dbInstance.DBInstanceIdentifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if len(dbInstances) == 0 {
		fmt.Fprintln(os.Stderr, fmt.Errorf("Not mutch RDS Instances Identifier %s", *dbInstance.DBInstanceIdentifier))
		return 1
	}

	endpoint := *dbInstances[0].Endpoint.Address

	authToken := dep.DNSimple.AuthToken
	accountId := dep.DNSimple.AccountID
	domain := dep.DNSimple.Domain
	recordId := dep.DNSimple.RecordID
	ttl := dep.DNSimple.TTL

	if authToken != "" && accountId != "" && domain != "" && recordId != 0 {
		if err = dnsimple.UpdateRecord(authToken, accountId, domain, recordId, endpoint, ttl); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	fmt.Println("Update DNS Record")

	prev := deployment.Previous{
		InstanceIdentifier: dep.Current.InstanceIdentifier,
		Endpoint:           dep.Current.Endpoint,
	}

	cur := deployment.Current{
		InstanceIdentifier: dep.Previous.InstanceIdentifier,
		Endpoint:           endpoint,
	}

	dep.Current = cur
	dep.Previous = prev
	if err = dep.Put(bucket, name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

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
