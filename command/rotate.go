package command

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/munisystem/rosculus/config"
	"github.com/munisystem/rosculus/database/rds"
	"github.com/munisystem/rosculus/dns/dnsimple"
	"github.com/munisystem/rosculus/lib/postgres"
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

	config, err := config.Load(bucket, name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	now := time.Now()
	dbInstanceIdentifier := fmt.Sprintf("%s-%s", config.DBInstanceIdentifierBase, now.Format("20060102"))
	prevDBInstanceIdentifier := fmt.Sprintf("%s-%s", config.DBInstanceIdentifierBase, now.Add(-24*time.Hour).Format("20060102"))

	dbInstanceConfig := &rds.DBInstanceConfig{
		SourceDBInstanceIdentifier: config.SourceDBInstanceIdentifier,
		TargetDBInstanceIdentifier: dbInstanceIdentifier,
		AvailabilityZone:           config.AvailabilityZone,
		PubliclyAccessible:         config.PubliclyAccessible,
		DBInstanceClass:            config.DBInstanceClass,
		DBSubnetGroupName:          config.DBSubnetGroupName,
		VpcSecurityGroupIds:        config.VPCSecurityGroupIds,
		Tags:                       config.DBInstanceTags,
		MasterUserPassword:         config.DBMasterUserPassword,
	}

	instance, err := rds.CloneDBInstance(dbInstanceConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if len(config.Queries) != 0 {
		connectionString := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s",
			instance.User,
			instance.Password,
			instance.URL,
			instance.Port,
			instance.Database,
		)
		p := postgres.Initialize(connectionString)

		if err := p.RunQueries(config.Queries); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	authToken := config.DNSimple.AuthToken
	accountId := config.DNSimple.AccountID
	domain := config.DNSimple.Domain
	recordName := config.DNSimple.RecordName
	ttl := config.DNSimple.TTL

	dnsClient := dnsimple.NewClient(authToken, accountId)
	if err := dnsClient.UpdateRecord(domain, recordName, instance.URL, ttl); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("Updated DNS Record")

	if err := rds.DeleteDBInstance(prevDBInstanceIdentifier); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("Failed to delete the previous DB Instance %s", prevDBInstanceIdentifier))
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
