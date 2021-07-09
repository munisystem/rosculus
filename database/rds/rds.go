package rds

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/rds"
	awspkg "github.com/munisystem/rosculus/aws"
	"github.com/munisystem/rosculus/database"
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

type DBInstanceConfig struct {
	SourceDBInstanceIdentifier string
	TargetDBInstanceIdentifier string
	AvailabilityZone           string
	PubliclyAccessible         bool
	DBInstanceClass            string
	DBSubnetGroupName          string
	VpcSecurityGroupIds        []string
	Tags                       map[string]string
	MasterUserPassword         string
}

func (config *DBInstanceConfig) tags() []*rds.Tag {
	rdsTags := make([]*rds.Tag, 0, len(config.Tags))
	for key, value := range config.Tags {
		rdsTags = append(rdsTags, &rds.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
	return rdsTags
}

func (config *DBInstanceConfig) vpcSecurityGroupIds() []*string {
	rdsVpcSecurityIds := make([]*string, len(config.VpcSecurityGroupIds), len(config.VpcSecurityGroupIds))
	for i := 0; i < len(config.VpcSecurityGroupIds); i++ {
		rdsVpcSecurityIds[i] = aws.String(config.VpcSecurityGroupIds[i])
	}
	return rdsVpcSecurityIds
}

func CloneDBInstance(config *DBInstanceConfig) (*database.DBInstance, error) {
	cli := client()

	var (
		instance *rds.DBInstance
		err      error
	)
	if instance, err = dbInstance(config.TargetDBInstanceIdentifier); err != nil {
		return nil, err
	} else if instance == nil {
		input := &rds.RestoreDBInstanceToPointInTimeInput{
			SourceDBInstanceIdentifier: aws.String(config.SourceDBInstanceIdentifier),
			TargetDBInstanceIdentifier: aws.String(config.TargetDBInstanceIdentifier),
			AvailabilityZone:           aws.String(config.AvailabilityZone),
			PubliclyAccessible:         aws.Bool(config.PubliclyAccessible),
			DBInstanceClass:            aws.String(config.DBInstanceClass),
			DBSubnetGroupName:          aws.String(config.DBSubnetGroupName),
			UseLatestRestorableTime:    aws.Bool(true),
			VpcSecurityGroupIds:        config.vpcSecurityGroupIds(),
			Tags:                       config.tags(),
		}
		resp, err := cli.RestoreDBInstanceToPointInTime(input)
		if err != nil {
			return nil, err
		}
		instance = resp.DBInstance
		log.Printf("created RDS Instance %s\n", config.TargetDBInstanceIdentifier)
	} else {
		log.Printf("RDS Instance %s is already exists\n", config.TargetDBInstanceIdentifier)
	}

	if err := waitUntilDBInstanceAvailable(config.TargetDBInstanceIdentifier); err != nil {
		return nil, err
	}

	if err := modifyDBInstance(config); err != nil {
		return nil, err
	}

	return &database.DBInstance{
		URL:      *instance.Endpoint.Address,
		Port:     *instance.Endpoint.Port,
		Database: *instance.DBName,
		User:     *instance.MasterUsername,
		Password: config.MasterUserPassword,
	}, nil
}

func modifyDBInstance(config *DBInstanceConfig) error {
	log.Printf("modify RDS Instance %s\n", config.TargetDBInstanceIdentifier)
	cli := client()

	input := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(config.TargetDBInstanceIdentifier),
		PubliclyAccessible:   aws.Bool(config.PubliclyAccessible),
		DBInstanceClass:      aws.String(config.DBInstanceClass),
		VpcSecurityGroupIds:  config.vpcSecurityGroupIds(),
		MasterUserPassword:   aws.String(config.MasterUserPassword),
		ApplyImmediately:     aws.Bool(true),
	}

	if _, err := cli.ModifyDBInstance(input); err != nil {
		return err
	}
	log.Printf("modified RDS Instance %s\n", config.TargetDBInstanceIdentifier)

	return waitUntilDBInstanceAvailable(config.TargetDBInstanceIdentifier)
}

func DeleteDBInstance(dbInstanceIdentifier string) error {
	cli := client()

	params := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		SkipFinalSnapshot:    aws.Bool(true),
	}
	if _, err := cli.DeleteDBInstance(params); err != nil {
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == rds.ErrCodeDBInstanceNotFoundFault {
				return nil
			}
			return err
		}
	}

	return nil
}

func waitUntilDBInstanceAvailable(dbInstanceIdentifier string) error {
	log.Printf("wait until RDS Instance %s is ready\n", dbInstanceIdentifier)
	cli := client()

	for {
		err := cli.WaitUntilDBInstanceAvailable(&rds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String(dbInstanceIdentifier)})
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == request.WaiterResourceNotReadyErrorCode {
				continue
			}
			return err
		}
		break
	}
	log.Printf("RDS Instance %s is ready\n", dbInstanceIdentifier)

	return nil
}

func dbInstance(dbInstanceIdentifier string) (*rds.DBInstance, error) {
	cli := client()

	resp, err := cli.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
	})
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == rds.ErrCodeDBInstanceNotFoundFault {
			return nil, nil
		}
		return nil, err
	}

	return resp.DBInstances[0], nil
}
