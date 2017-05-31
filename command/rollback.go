package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/munisystem/rstack/aws/rds"
	"github.com/munisystem/rstack/deployment"
	"github.com/munisystem/rstack/dnsimple"
)

type RollbackCommand struct {
	Meta
}

func (c *RollbackCommand) Run(args []string) int {
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

	exists, err := rds.DBInstanceAllreadyExists(dep.Previous.InstanceIdentifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if exists != true {
		fmt.Fprintln(os.Stderr, errors.New("Previous DB Instance is not exists"))
		return 1
	}

	dbInstances, err := rds.DescribeDBInstances(dep.Previous.InstanceIdentifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if *dbInstances[0].DBInstanceStatus != "available" {
		fmt.Fprintln(os.Stderr, errors.New("Previous DB Instance is not available"))
		return 1
	}

	endpoint := *dbInstances[0].Endpoint.Address
	authToken := dep.DNSimple.AuthToken
	accountId := dep.DNSimple.AccountID
	domain := dep.DNSimple.Domain
	recordId := dep.DNSimple.RecordID
	recordName := dep.DNSimple.RecordName
	ttl := dep.DNSimple.TTL

	if authToken == "" && accountId == "" && domain == "" && recordId == 0 && recordName == "" {
		fmt.Fprintln(os.Stderr, errors.New("Lack DNSimple configs, change only statuses"))
	} else {

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

func (c *RollbackCommand) Synopsis() string {
	return ""
}

func (c *RollbackCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
