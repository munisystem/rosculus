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

	dep, err := deployment.Load(bucket, name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// FIXME: Will remove information about current instance identifier
	baseIdentifier := dep.Current.InstanceIdentifier

	var oldInstanceIdentifier, newInstanceIdentifier string
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	today := time.Now().In(loc)
	yesterday := today.AddDate(0, 0, -1)
	newInstanceIdentifier = baseIdentifier + "-" + today.Format("20060102")
	oldInstanceIdentifier = baseIdentifier + "-" + yesterday.Format("20060102")
	if oldInstanceIdentifier == newInstanceIdentifier {
		oldInstanceIdentifier = ""
	}

	exist, err := rds.DBInstanceAlreadyExists(newInstanceIdentifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if exist == true {
		fmt.Printf("'%s' is already exists, delete before to launch new DB instance\n", newInstanceIdentifier)
		prevDBInstances, err := rds.DescribeDBInstances(newInstanceIdentifier)
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
			fmt.Println("Deleted DB instance")
		}
	}

	fmt.Printf("Launch new RDS Instance '%s' from '%s'\n", newInstanceIdentifier, dep.SourceDBInstanceIdentifier)
	dbInstance, err := rds.CopyInstance(
		dep.SourceDBInstanceIdentifier,
		newInstanceIdentifier,
		dep.AvailabilityZone,
		dep.DBInstanceClass,
		dep.DBSubnetGroupName,
		dep.PubliclyAccessible,
		dep.DBInstanceTags,
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
	port := *dbInstances[0].Endpoint.Port
	user := *dbInstances[0].MasterUsername
	database := *dbInstances[0].DBName

	if len(dep.Queries) != 0 {
		connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, dep.DBMasterUserPassword, endpoint, port, database)
		p := postgres.Initialize(connectionString)

		if err := p.RunQueries(dep.Queries); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

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

	if oldInstanceIdentifier != "" {
		fmt.Printf("delete previous DB Instance '%s'\n", oldInstanceIdentifier)
		prevDBInstances, err := rds.DescribeDBInstances(oldInstanceIdentifier)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		if len(prevDBInstances) != 0 {
			if *prevDBInstances[0].DBInstanceStatus != "deleting" {
				tags, err := rds.GetInstanceTags(*prevDBInstances[0].DBInstanceArn)
				if err != nil {
					fmt.Printf("failed to get previous DB Instances tags, skip deleting '%s'\n", err)
					return 0
				}
				for _, tag := range tags {
					if *tag.Key != "env" {
						continue
					}
					if *tag.Value == "production" {
						fmt.Printf("'%s' is for production environment, skip deleting\n", oldInstanceIdentifier)
						return 0
					}
				}
				if _, err := rds.DeleteInstance(*prevDBInstances[0].DBInstanceIdentifier); err != nil {
					fmt.Fprintln(os.Stderr, err)
					return 1
				}
			}
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
