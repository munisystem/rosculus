package dnsimple

import (
	"context"

	"github.com/dnsimple/dnsimple-go/dnsimple"
)

func UpdateRecord(token, accountId, domain string, recordId int, recordName, content string, ttl int) error {
	tc := dnsimple.StaticTokenHTTPClient(context.Background(), token)
	client := dnsimple.NewClient(tc)

	attributes := &dnsimple.ZoneRecordAttributes{
		Name:    dnsimple.String(recordName),
		Content: content,
		TTL:     ttl,
	}

	if _, err := client.Zones.UpdateRecord(context.Background(), accountId, domain, int64(recordId), *attributes); err != nil {
		return err
	}

	return nil
}
