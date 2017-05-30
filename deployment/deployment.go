package deployment

import (
	"github.com/munisystem/rstack/aws/s3"
	yaml "gopkg.in/yaml.v2"
)

type Deployment struct {
	SourceDBInstanceIdentifier string   `yaml:"SourceDBInstanceIdentifier"`
	PubliclyAccessible         bool     `yaml:"PubliclyAccessible"`
	DBInstanceClass            string   `yaml:"DBInstanceClass"`
	SecurityGroups             []string `yaml:"SecurityGroups"`
	DNSimple                   DNSimple `yaml:"DNSimple"`
	Current                    Current  `yaml:"Current"`
	Previous                   Previous `yaml:"Previous"`
}

type DNSimple struct {
	AuthToken string `yaml:"AuthToken"`
	AccountID string `yaml:"AccountID"`
	Domain    string `yaml:"Domain"`
	RecordID  int    `yaml:"RecordID"`
	TTL       int    `yaml:"TTL"`
}

type Current struct {
	InstanceIdentifier string `yaml:"InstanceIdentifier"`
	Endpoint           string `yaml:"Endpoint"`
}

type Previous struct {
	InstanceIdentifier string `yaml:"InstanceIdentifier"`
	Endpoint           string `yaml:"Endpoint"`
}

func Load(bucket, name string) (*Deployment, error) {
	c := &Deployment{}

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

func (dep *Deployment) Put(bucket, name string) error {
	str, err := yaml.Marshal(dep)
	if err != nil {
		return err
	}

	key := name + ".yml"
	if err = s3.Upload(bucket, key, []byte(str)); err != nil {
		return err
	}

	return nil
}
