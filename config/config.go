package config

import (
	"github.com/munisystem/rosculus/aws/s3"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	// SourceDBClusterIdentifier  string            `yaml:"SourceDBClusterIdentifier,omitempty"`
	SourceDBInstanceIdentifier string            `yaml:"SourceDBInstanceIdentifier"`
	DBInstanceIdentifierBase   string            `yaml:"DBInstanceIdentifierBase"`
	DBMasterUserPassword       string            `yaml:"DBMasterUserPassword"`
	DBInstanceTags             map[string]string `yaml:"DBInstanceTags"`
	AvailabilityZone           string            `yaml:"AvailabilityZone"`
	DBSubnetGroupName          string            `yaml:"DBSubnetGroupName"`
	PubliclyAccessible         bool              `yaml:"PubliclyAccessible"`
	DBInstanceClass            string            `yaml:"DBInstanceClass"`
	VPCSecurityGroupIds        []string          `yaml:"VPCSecurityGroupIds"`
	DNSimple                   DNSimple          `yaml:"DNSimple"`
	Rollback                   bool              `yaml:"Rollback"`
	Queries                    []string          `yaml:"Queries"`
}

type DNSimple struct {
	AuthToken  string `yaml:"AuthToken"`
	AccountID  string `yaml:"AccountID"`
	Domain     string `yaml:"Domain"`
	RecordID   int    `yaml:"RecordID"`
	RecordName string `yaml:"RecordName"`
	TTL        int    `yaml:"TTL"`
}

func Load(bucket, name string) (*Config, error) {
	c := &Config{}

	key := name + ".yml"
	buf, err := s3.Download(bucket, key)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
