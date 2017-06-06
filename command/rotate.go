package command

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/munisystem/rosculus/aws/rds"
	"github.com/munisystem/rosculus/deployment"
	"github.com/munisystem/rosculus/dnsimple"
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

	exist, err := rds.DBInstanceAllreadyExists(dep.Previous.InstanceIdentifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if exist == true {
		fmt.Printf("'%s' is already exists, delete before to launch new DB instance\n", dep.Previous.InstanceIdentifier)
		prevDBInstances, err := rds.DescribeDBInstances(dep.Previous.InstanceIdentifier)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		if len(prevDBInstances) != 0 {
			if *prevDBInstances[0].DBInstanceStatus != "deleting" {
				if _, err := rds.DeleteInstance(*prevDBInstances[0].DBInstanceIdentifier); err != nil {
					fmt.Fprintln(os.Stderr, err)
					return 1
				}
			}

			if err = wait(rds.WaitDeleted, *prevDBInstances[0].DBInstanceIdentifier); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}

			fmt.Println()
			fmt.Println("Deleted previous DB instance")
		}
	}

	fmt.Printf("Launch new RDS Instance '%s' from '%s'\n", dep.Previous.InstanceIdentifier, dep.SourceDBInstanceIdentifier)
	dbInstance, err := rds.CopyInstance(
		dep.SourceDBInstanceIdentifier,
		dep.Previous.InstanceIdentifier,
		dep.AvailabilityZone,
		dep.DBInstanceClass,
		dep.DBSubnetGroupName,
		dep.PubliclyAccessible,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("Launched. Please Wait RDS Instance ready")

	if err = wait(rds.WaitReady, *dbInstance.DBInstanceIdentifier); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Println()
	fmt.Printf("%s is ready\n", *dbInstance.DBInstanceIdentifier)

	if err = rds.AddSecurityGroups(*dbInstance.DBInstanceIdentifier, dep.VPCSecurityGroupIds); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Println("Attached security groups", dep.VPCSecurityGroupIds)

	if dep.DBMasterUserPassword != "" {
		if err = rds.ChangeMasterPassword(*dbInstance.DBInstanceIdentifier, dep.DBMasterUserPassword); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println("Changed Master User Password")
	}

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
	recordName := dep.DNSimple.RecordName
	ttl := dep.DNSimple.TTL

	if authToken != "" && accountId != "" && domain != "" && recordId != 0 && recordName != "" {
		if err = dnsimple.UpdateRecord(authToken, accountId, domain, recordId, recordName, endpoint, ttl); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println("Updated DNS Record")
	}

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

func wait(fn func(string) error, str string) error {
	errCh := make(chan error, 2)
	defer close(errCh)

	go func() {
		errCh <- fn(str)
	}()

	go func() {
	loop:
		for {
			if len(errCh) > 0 {
				break loop
			}

			fmt.Print(".")
			time.Sleep(30 * time.Second)
		}

	}()

	err := <-errCh

	return err
}
