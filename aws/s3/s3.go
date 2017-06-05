package s3

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	awspkg "github.com/munisystem/rosculus/aws"
)

var (
	s3cli *s3.S3
)

func client() *s3.S3 {
	if s3cli == nil {
		s3cli = s3.New(awspkg.Session())
	}
	return s3cli
}

func Download(bucket, key string) ([]byte, error) {
	cli := client()

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	resp, err := cli.GetObject(params)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, resp.Body)

	return buf.Bytes(), nil
}

func Upload(bucket, key string, body []byte) error {
	cli := client()

	params := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	}

	if _, err := cli.PutObject(params); err != nil {
		return err
	}

	return nil
}
