package deployment

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
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

func New() {
}
