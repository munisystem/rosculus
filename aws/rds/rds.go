package rds

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/rds"
	awspkg "github.com/munisystem/rosculus/aws"
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

func CopyInstance(sourceDBInstanceIdentifier, targetDBInstanceIdentifier, availabilityZone, dbInstanceClass, dbSubnetGroupName string, publiclyAccessible bool, instanceTags map[string]string) (*rds.DBInstance, error) {
	cli := client()

	var tags []*rds.Tag
	if len(instanceTags) > 0 {
		for key, value := range instanceTags {
			tags = append(tags, &rds.Tag{Key: aws.String(key), Value: aws.String(value)})
		}
	}

	params := &rds.RestoreDBInstanceToPointInTimeInput{
		SourceDBInstanceIdentifier: aws.String(sourceDBInstanceIdentifier),
		TargetDBInstanceIdentifier: aws.String(targetDBInstanceIdentifier),
		AvailabilityZone:           aws.String(availabilityZone),
		PubliclyAccessible:         aws.Bool(publiclyAccessible),
		DBInstanceClass:            aws.String(dbInstanceClass),
		DBSubnetGroupName:          aws.String(dbSubnetGroupName),
		UseLatestRestorableTime:    aws.Bool(true),
		Tags: tags,
	}

	resp, err := cli.RestoreDBInstanceToPointInTime(params)
	if err != nil {
		return nil, err
	}

	return resp.DBInstance, nil
}

func DeleteInstance(dbInstanceIdentifier string) (*rds.DBInstance, error) {
	cli := client()

	params := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		SkipFinalSnapshot:    aws.Bool(true),
	}

	resp, err := cli.DeleteDBInstance(params)
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

	var err error
	for {
		err = cli.WaitUntilDBInstanceAvailable(params)
		if !checkWaiterNotReady(err) {
			break
		}
	}

	return err
}

func checkWaiterNotReady(err error) bool {
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == request.WaiterResourceNotReadyErrorCode {
			return true
		}
		return false
	}

	return false
}

func WaitDeleted(dbInstanceIdentifier string) error {
	cli := client()

	params := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
	}

	if err := cli.WaitUntilDBInstanceDeleted(params); err != nil {
		return err
	}

	return nil
}

func DBInstanceAllreadyExists(dbInstanceIdentifier string) (bool, error) {
	cli := client()

	resp, err := cli.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		return false, err
	}

	exist := false
	if len(resp.DBInstances) > 0 {
		for _, v := range resp.DBInstances {
			if dbInstanceIdentifier == *v.DBInstanceIdentifier {
				exist = true
				break
			}
		}
	}

	return exist, nil
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

func AddSecurityGroups(dbInstanceIdentifier string, vpcSecurityGroupIds []string) error {
	cli := client()

	var awsVPCSecurityGroupIds []*string
	for _, v := range vpcSecurityGroupIds {
		awsVPCSecurityGroupIds = append(awsVPCSecurityGroupIds, aws.String(v))
	}

	params := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		VpcSecurityGroupIds:  awsVPCSecurityGroupIds,
	}

	if _, err := cli.ModifyDBInstance(params); err != nil {

		return err
	}

	return nil
}

func ChangeMasterPassword(dbInstanceIdentifier, password string) error {
	cli := client()

	params := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		MasterUserPassword:   aws.String(password),
	}

	if _, err := cli.ModifyDBInstance(params); err != nil {
		return err
	}

	return nil
}
