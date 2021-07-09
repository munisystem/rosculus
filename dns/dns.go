package dns

type DNS interface {
	UpdateRecord(domain, name, value string, ttl int) error
}
