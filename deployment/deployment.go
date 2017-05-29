package deployment

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
	"github.com/munisystem/rstack/aws/s3"
)

type Deployment struct {
	SourceDBInstanceIdentifier string `yaml:"SourceDBInstanceIdentifier"`
	PubliclyAccessible bool `yaml:"PubliclyAccessible"`
	DBInstanceClass string `yaml:"DBInstanceClass"`
	Current Current `yaml:"Current"`
	Previous Previous `yaml:"Previous"`
}

type Current struct {
	InstanceIdentifier string `yaml:"InstanceIdentifier"`
	Endpoint string `yaml:"Endpoint"`
}

type Previous struct {
	InstanceIdentifier string `yaml:"InstanceIdentifier"`
	Endpoint string `yaml:"Endpoint"`
}

func Load(deployment string) (*Deployment, error) {
	c := &Deployment{}

	src := deployment + ".yml"
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, c)

	return c, nil;
}

func (dep *Deployment) New(bucket, name string) error {
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
