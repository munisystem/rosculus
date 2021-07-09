package database

type DBInstance struct {
	URL      string
	Port     int64
	Database string
	User     string
	Password string
}
