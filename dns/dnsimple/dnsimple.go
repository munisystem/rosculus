package dnsimple

import (
	"context"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/munisystem/rosculus/dns"
)

type Client struct {
	client    *dnsimple.Client
	accountID string
}

func NewClient(token, accountID string) dns.DNS {
	tc := dnsimple.StaticTokenHTTPClient(context.Background(), token)
	client := dnsimple.NewClient(tc)

	return &Client{
		client:    client,
		accountID: accountID,
	}
}

func (c *Client) UpdateRecord(domain, name, value string, ttl int) error {
	ctx := context.Background()
	if recordID, err := c.getRecordID(ctx, domain, name); err != nil {
		return err
	} else if recordID == -1 {
		return c.createRecord(ctx, domain, name, value, ttl)
	} else {
		attributes := &dnsimple.ZoneRecordAttributes{
			Name:    dnsimple.String(name),
			Content: value,
			TTL:     ttl,
		}
		if _, err := c.client.Zones.UpdateRecord(context.Background(), c.accountID, domain, recordID, *attributes); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) getRecordID(ctx context.Context, domain, name string) (int64, error) {
	options := &dnsimple.ZoneRecordListOptions{
		Name: dnsimple.String(name),
		Type: dnsimple.String("CNAME"),
	}
	resp, err := c.client.Zones.ListRecords(ctx, c.accountID, domain, options)
	if err != nil {
		return -1, err
	}
	if len(resp.Data) == 0 {
		return -1, nil
	}
	return resp.Data[0].ID, nil
}

func (c *Client) createRecord(ctx context.Context, domain, name, value string, ttl int) error {
	attributes := &dnsimple.ZoneRecordAttributes{
		Name:    dnsimple.String(name),
		Type:    "CNAME",
		Content: value,
		TTL:     ttl,
	}
	if _, err := c.client.Zones.CreateRecord(ctx, c.accountID, domain, *attributes); err != nil {
		return err
	}

	return nil
}
