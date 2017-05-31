package dnsimple

import "github.com/dnsimple/dnsimple-go/dnsimple"

func UpdateRecord(token, accountId, domain string, recordId int, recordName, content string, ttl int) error {
	client := dnsimple.NewClient(dnsimple.NewOauthTokenCredentials(token))

	attributes := &dnsimple.ZoneRecord{
		Name:    recordName,
		Content: content,
		TTL:     ttl,
	}

	if _, err := client.Zones.UpdateRecord(accountId, domain, recordId, *attributes); err != nil {
		return err
	}

	return nil
}
