package postgres

type ClientOptions struct {
	Host     string
	Port     int
	Databse  string
	Username string
	Password string
	SslMode  string
}

func NewClientOptions() *ClientOptions {
	return &ClientOptions{
		Host:     "localhost",
		Port:     5432,
		Databse:  "postgres",
		Username: "postgres",
		Password: "postgres",
		SslMode:  "disable",
	}
}

func (o *ClientOptions) SetHost(host string) *ClientOptions {
	o.Host = host
	return o
}

func (o *ClientOptions) SetPort(port int) *ClientOptions {
	o.Port = port
	return o
}

func (o *ClientOptions) SetDatabase(database string) *ClientOptions {
	o.Databse = database
	return o
}

func (o *ClientOptions) SetUsername(username string) *ClientOptions {
	o.Username = username
	return o
}

func (o *ClientOptions) SetPassword(password string) *ClientOptions {
	o.Password = password
	return o
}
