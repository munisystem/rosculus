package rds

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	awspkg "github.com/munisystem/rstack/aws"
)

var (
	rdscli *rds.RDS
)

func client() *rds.RDS {
	if rdscli == nil {
		rdscli = rds.New(awspkg.Session())
	}
	return rdscli
}

func CopyInstance(sourceDBInstanceIdentifier, targetDBInstanceIdentifier, dbInstanceClass string, publiclyAccessible bool) (*rds.DBInstance, error) {
	cli := client()

	params := &rds.RestoreDBInstanceToPointInTimeInput{
		SourceDBInstanceIdentifier: aws.String(sourceDBInstanceIdentifier),
		TargetDBInstanceIdentifier: aws.String(targetDBInstanceIdentifier),
		PubliclyAccessible:         aws.Bool(publiclyAccessible),
		DBInstanceClass:            aws.String(dbInstanceClass),
		UseLatestRestorableTime:    aws.Bool(true),
	}

	resp, err := cli.RestoreDBInstanceToPointInTime(params)
	if err != nil {
		return nil, err
	}

	return resp.DBInstance, nil
}

func WaitReady(dbInstanceIdentifier string) error {
	cli := client()

	params := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
	}

	err := cli.WaitUntilDBInstanceAvailable(params)
	if err != nil {
		return err
	}
	return nil
}

func DescribeDBInstances(dbInstanceIdentifier string) ([]*rds.DBInstance, error) {
	cli := client()

	params := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
	}

	resp, err := cli.DescribeDBInstances(params)
	if err != nil {
		return nil, err
	}

	return resp.DBInstances, nil
}
