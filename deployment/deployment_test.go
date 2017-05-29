package deployment

import (
	"os"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	expected := &Deployment{
		SourceDBInstanceIdentifier: "cat",
		PubliclyAccessible: true,
		DBInstanceClass: "db.d3.medium",
		Current: Current{
			InstanceIdentifier: "cat-blue",
			Endpoint: "cat-blue.abcdefghijklmn.ap-northeast-1.rds.amazonaws.com",
		},
		Previous: Previous{
			InstanceIdentifier: "cat-green",
			Endpoint: "cat-green.abcdefghijklmn.ap-northeast-1.rds.amazonaws.com",
		},
	}

	dir, _ := os.Getwd()
	deployment := dir + "/" + "example"

	actual, err := Load(deployment)
	if err != nil {
		t.Errorf("Error should not be railed. error: %s", err)
	}

	if reflect.DeepEqual(actual, expected) {
		t.Errorf("Config does not match. expected: %q, actual: %q", expected, actual)
	}

}
