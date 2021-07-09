package command

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/munisystem/rosculus/config"
	"github.com/munisystem/rosculus/database"
	"github.com/munisystem/rosculus/database/rds"
	"github.com/munisystem/rosculus/dns/dnsimple"
	"github.com/munisystem/rosculus/lib/postgres"
)

type RotateCommand struct {
	Meta
}

func (c *RotateCommand) Run(args []string) int {
	if len(args) == 0 {
		log.Fatalln("too few arguments")
	} else if len(args) > 1 {
		log.Fatalln("too many arguments")
	}

	name := args[0]

	bucket := os.Getenv("AWS_S3_BUCKET_NAME")
	if bucket == "" {
		log.Fatalln("please set s3 bucket name in AWS_S3_BUCKET_NAME")
	}

	config, err := config.Load(bucket, name)
	if err != nil {
		log.Fatalf("failed to load config file from S3: %s\n", err)
	}

	var (
		dbIdentifier     string
		prevDBIdentifier string
		instance         *database.DBInstance
	)
	now := time.Now()

	if config.SourceDBInstanceIdentifier != "" && config.DBInstanceIdentifier != "" {
		dbIdentifier = fmt.Sprintf("%s-%s", config.DBInstanceIdentifier, now.Format("20060102"))
		prevDBIdentifier = fmt.Sprintf("%s-%s", config.DBInstanceIdentifier, now.Add(-24*time.Hour).Format("20060102"))

		dbInstanceConfig := &rds.DBInstanceConfig{
			SourceDBInstanceIdentifier: config.SourceDBInstanceIdentifier,
			TargetDBInstanceIdentifier: dbIdentifier,
			AvailabilityZone:           config.AvailabilityZone,
			PubliclyAccessible:         config.PubliclyAccessible,
			DBInstanceClass:            config.DBInstanceClass,
			DBSubnetGroupName:          config.DBSubnetGroupName,
			VpcSecurityGroupIds:        config.VPCSecurityGroupIds,
			Tags:                       config.DBInstanceTags,
			MasterUserPassword:         config.DBMasterUserPassword,
		}

		instance, err = rds.CloneDBInstance(dbInstanceConfig)
	} else if config.SourceDBClusterIdentifier != "" && config.DBClusterIdentifier != "" {
		dbIdentifier = fmt.Sprintf("%s-%s", config.DBClusterIdentifier, now.Format("20060102"))
		prevDBIdentifier = fmt.Sprintf("%s-%s", config.DBClusterIdentifier, now.Add(-24*time.Hour).Format("20060102"))

		dbClusterConfig := &rds.DBClusterConfig{
			SourceDBClusterIdentifier: config.SourceDBClusterIdentifier,
			DBClusterIdentifier:       dbIdentifier,
			AvailabilityZone:          config.AvailabilityZone,
			PubliclyAccessible:        config.PubliclyAccessible,
			DBInstanceClass:           config.DBInstanceClass,
			DBSubnetGroupName:         config.DBSubnetGroupName,
			VpcSecurityGroupIds:       config.VPCSecurityGroupIds,
			Tags:                      config.DBInstanceTags,
			MasterUserPassword:        config.DBMasterUserPassword,
		}

		instance, err = rds.CloneDBCluster(dbClusterConfig)
	} else {
		log.Fatalf("config %s is invalid\n", name)
	}

	if err != nil {
		log.Fatalf("failed to create Database: %s\n", err)
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
			log.Fatalf("failed to execute queries: %s\n", err)
		}

		log.Println("executed queries")
	}

	authToken := config.DNSimple.AuthToken
	accountId := config.DNSimple.AccountID
	domain := config.DNSimple.Domain
	recordName := config.DNSimple.RecordName
	ttl := config.DNSimple.TTL

	dnsClient := dnsimple.NewClient(authToken, accountId)
	if err := dnsClient.UpdateRecord(domain, recordName, instance.URL, ttl); err != nil {
		log.Fatalf("failed to update DNS record %s: %s \n", recordName, err)
	}
	log.Printf("updated DNS record %s.%s\n", recordName, domain)

	if config.SourceDBInstanceIdentifier != "" && config.DBInstanceIdentifier != "" {
		if err := rds.DeleteDBInstance(prevDBIdentifier); err != nil {
			log.Fatalf("failed to delete the previous DB Instance %s: %s\n", prevDBIdentifier, err)
		}
	} else if config.SourceDBClusterIdentifier != "" && config.DBClusterIdentifier != "" {
		if err := rds.DeleteDBCluster(prevDBIdentifier); err != nil {
			log.Fatalf("failed to delete the previous DB Cluster %s: %s\n", prevDBIdentifier, err)
		}
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
