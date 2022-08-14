package postgres

import (
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strconv"
)

type Client interface {
	// Connect to Postgres
	Connect() error
	// Disconnect from Postgres
	Disconnect() error

	Migrate() error
	Execute(query string) error
	Select(query string) interface{}
}

type client struct {
	//mqttClient mqtt.Client
	db      *sql.DB
	options ClientOptions
	log     zerolog.Logger
}

func NewClient(options *ClientOptions) Client {
	logger := log.With().Str("Component", "Postgres").Logger()

	return &client{
		options: *options,
		log:     logger,
	}
}

func (c client) Connect() error {
	c.log.Info().Str("host", c.options.Host).Int("port", c.options.Port).Str("database", c.options.Databse).Msg("Connecting to database")
	connStr := "postgres://" + c.options.Username + ":" + c.options.Password + "@" + c.options.Host + ":" + strconv.Itoa(c.options.Port) + "/" + c.options.Databse + "?sslmode=enable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.log.Error().Err(err).Msg("Unable to connect")
		return err
	}
	if db == nil {
		log.Error().Msg("db is nil")
	}
	c.db = db
	return nil
}

func (c client) Disconnect() error {
	return nil
}

func (c client) Execute(query string) error {
	//TODO implement me
	panic("implement me")
}

func (c client) Migrate() error {
	driver, err := postgresMigrate.WithInstance(c.db, &postgresMigrate.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"postgres", driver)
	err = m.Up()
	if err != nil {
		return err
	}
	return nil
}

func (c client) Select(query string) interface{} {
	//TODO implement me
	panic("implement me")
}
