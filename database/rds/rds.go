package rds

import (
	"fmt"
	"log"
	"time"

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

func tags(tags map[string]string) []*rds.Tag {
	rdsTags := make([]*rds.Tag, 0, len(tags))
	for key, value := range tags {
		rdsTags = append(rdsTags, &rds.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
	return rdsTags
}

func vpcSecurityGroupIds(ids []string) []*string {
	rdsVpcSecurityIds := make([]*string, len(ids), len(ids))
	for i := 0; i < len(ids); i++ {
		rdsVpcSecurityIds[i] = aws.String(ids[i])
	}
	return rdsVpcSecurityIds
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

func CloneDBInstance(config *DBInstanceConfig) (*database.DBInstance, error) {
	cli := client()

	if instance, err := dbInstance(config.TargetDBInstanceIdentifier); err != nil {
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
			VpcSecurityGroupIds:        vpcSecurityGroupIds(config.VpcSecurityGroupIds),
			Tags:                       tags(config.Tags),
		}
		if _, err := cli.RestoreDBInstanceToPointInTime(input); err != nil {
			return nil, err
		}
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

	instance, err := dbInstance(config.TargetDBInstanceIdentifier)
	if err != nil {
		return nil, err
	} else if instance == nil {
		return nil, fmt.Errorf("failed to get informations of RDS Instance %s", config.TargetDBInstanceIdentifier)
	}

	return &database.DBInstance{
		URL:      *instance.Endpoint.Address,
		Port:     *instance.Endpoint.Port,
		Database: *instance.DBName,
		User:     *instance.MasterUsername,
		Password: config.MasterUserPassword,
	}, nil
}

type DBClusterConfig struct {
	SourceDBClusterIdentifier string
	DBClusterIdentifier       string
	AvailabilityZone          string
	PubliclyAccessible        bool
	DBInstanceClass           string
	DBSubnetGroupName         string
	VpcSecurityGroupIds       []string
	Tags                      map[string]string
	MasterUserPassword        string
}

func CloneDBCluster(config *DBClusterConfig) (*database.DBInstance, error) {
	cli := client()

	if cluster, err := dbCluster(config.DBClusterIdentifier); err != nil {
		return nil, err
	} else if cluster == nil {
		input := &rds.RestoreDBClusterToPointInTimeInput{
			SourceDBClusterIdentifier: aws.String(config.SourceDBClusterIdentifier),
			DBClusterIdentifier:       aws.String(config.DBClusterIdentifier),
			DBSubnetGroupName:         aws.String(config.DBSubnetGroupName),
			UseLatestRestorableTime:   aws.Bool(true),
			VpcSecurityGroupIds:       vpcSecurityGroupIds(config.VpcSecurityGroupIds),
			Tags:                      tags(config.Tags),
		}
		if _, err := cli.RestoreDBClusterToPointInTime(input); err != nil {
			return nil, err
		}
		log.Printf("created Aurora Cluster %s\n", config.DBClusterIdentifier)
	} else {
		log.Printf("Aurora Cluster %s is already exists\n", config.DBClusterIdentifier)
	}

	if err := waitUntilDBClusterAvailable(config.DBClusterIdentifier); err != nil {
		return nil, err
	}

	if err := modifyDBCluster(config); err != nil {
		return nil, err
	}

	if err := addDBInstanceToCluster(config); err != nil {
		return nil, err
	}

	cluster, err := dbCluster(config.DBClusterIdentifier)
	if err != nil {
		return nil, err
	} else if cluster == nil {
		return nil, fmt.Errorf("failed to get informations of Aurora cluster %s", config.DBClusterIdentifier)
	}

	return &database.DBInstance{
		URL:      *cluster.Endpoint,
		Port:     *cluster.Port,
		Database: *cluster.DatabaseName,
		User:     *cluster.MasterUsername,
		Password: config.MasterUserPassword,
	}, nil
}

func addDBInstanceToCluster(config *DBClusterConfig) error {
	cli := client()

	instanceIdentifier := config.DBClusterIdentifier + "-001"

	var (
		instance *rds.DBInstance
		err      error
	)

	if instance, err = dbInstance(instanceIdentifier); err != nil {
		return err
	} else if instance == nil {
		input := &rds.CreateDBInstanceInput{
			DBClusterIdentifier:  aws.String(config.DBClusterIdentifier),
			DBInstanceIdentifier: aws.String(instanceIdentifier),
			AvailabilityZone:     aws.String(config.AvailabilityZone),
			PubliclyAccessible:   aws.Bool(config.PubliclyAccessible),
			DBInstanceClass:      aws.String(config.DBInstanceClass),
			Engine:               aws.String("aurora-postgresql"),
			Tags:                 tags(config.Tags),
		}
		resp, err := cli.CreateDBInstance(input)
		if err != nil {
			return err
		}
		instance = resp.DBInstance
		log.Printf("created RDS Instance to Aurora Cluster %s\n", instanceIdentifier)
	} else {
		log.Printf("RDS Instance %s is already exists in Aurora Cluster %s\n", instanceIdentifier, config.DBClusterIdentifier)
	}

	if err := waitUntilDBInstanceAvailable(instanceIdentifier); err != nil {
		return err
	}

	return nil
}

func modifyDBInstance(config *DBInstanceConfig) error {
	log.Printf("modify RDS Instance %s\n", config.TargetDBInstanceIdentifier)
	cli := client()

	input := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: aws.String(config.TargetDBInstanceIdentifier),
		PubliclyAccessible:   aws.Bool(config.PubliclyAccessible),
		DBInstanceClass:      aws.String(config.DBInstanceClass),
		VpcSecurityGroupIds:  vpcSecurityGroupIds(config.VpcSecurityGroupIds),
		MasterUserPassword:   aws.String(config.MasterUserPassword),
		ApplyImmediately:     aws.Bool(true),
	}

	if _, err := cli.ModifyDBInstance(input); err != nil {
		return err
	}
	log.Printf("modified RDS Instance %s\n", config.TargetDBInstanceIdentifier)

	return waitUntilDBInstanceAvailable(config.TargetDBInstanceIdentifier)
}

func modifyDBCluster(config *DBClusterConfig) error {
	log.Printf("modify Aurora Cluster %s\n", config.DBClusterIdentifier)
	cli := client()

	input := &rds.ModifyDBClusterInput{
		DBClusterIdentifier: aws.String(config.DBClusterIdentifier),
		VpcSecurityGroupIds: vpcSecurityGroupIds(config.VpcSecurityGroupIds),
		MasterUserPassword:  aws.String(config.MasterUserPassword),
		ApplyImmediately:    aws.Bool(true),
	}

	if _, err := cli.ModifyDBCluster(input); err != nil {
		return err
	}
	log.Printf("modified Aurora Cluster %s\n", config.DBClusterIdentifier)

	return waitUntilDBClusterAvailable(config.DBClusterIdentifier)
}

func DeleteDBInstance(dbInstanceIdentifier string) error {
	cli := client()

	input := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
		SkipFinalSnapshot:    aws.Bool(true),
	}
	if _, err := cli.DeleteDBInstance(input); err != nil {
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == rds.ErrCodeDBInstanceNotFoundFault {
				return nil
			}
			return err
		}
	}

	return nil
}

func DeleteDBCluster(dbClusterIdentifier string) error {
	cli := client()

	input := &rds.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(dbClusterIdentifier),
		SkipFinalSnapshot:   aws.Bool(true),
	}

	resp, err := cli.DescribeDBClusters(&rds.DescribeDBClustersInput{DBClusterIdentifier: aws.String(dbClusterIdentifier)})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == rds.ErrCodeDBClusterNotFoundFault {
			return nil
		}
		return err
	}
	for _, member := range resp.DBClusters[0].DBClusterMembers {
		if err := DeleteDBInstance(*member.DBInstanceIdentifier); err != nil {
			return err
		}
	}
	if _, err := cli.DeleteDBCluster(input); err != nil {
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == rds.ErrCodeDBClusterNotFoundFault {
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

func waitUntilDBClusterAvailable(dbClusterIdentifier string) error {
	log.Printf("wait until Aurora Cluster %s is ready\n", dbClusterIdentifier)
	cli := client()

	maxAttempt := 120
	for i := 0; i < maxAttempt; i++ {
		resp, err := cli.DescribeDBClusters(&rds.DescribeDBClustersInput{DBClusterIdentifier: aws.String(dbClusterIdentifier)})
		if err != nil {
			return err
		}
		if *resp.DBClusters[0].Status == "available" {
			log.Printf("Aurora Cluster %s is ready\n", dbClusterIdentifier)
			return nil
		}
		time.Sleep(30 * time.Second)
	}

	return fmt.Errorf("Aurora Cluster %s is not ready, exceed max wait attemps", dbClusterIdentifier)
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

func dbCluster(dbClusterIdentifier string) (*rds.DBCluster, error) {
	cli := client()

	resp, err := cli.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(dbClusterIdentifier),
	})
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == rds.ErrCodeDBClusterNotFoundFault {
			return nil, nil
		}
		return nil, err
	}

	return resp.DBClusters[0], nil
}
