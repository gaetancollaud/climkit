package postgres

import (
	"database/sql"
	"fmt"
	"github.com/gaetancollaud/climkit-to-mqtt/pkg/postgres/migrations"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
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
	Execute(query string, args ...any) error
	Select(query string, args ...any) *sql.Row
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
		db:      nil,
	}
}

func (c *client) Connect() error {
	c.log.Info().Str("host", c.options.Host).Int("port", c.options.Port).Str("database", c.options.Databse).Msg("Connecting to database")
	connStr := "postgres://" + c.options.Username + ":" + c.options.Password + "@" + c.options.Host + ":" + strconv.Itoa(c.options.Port) + "/" + c.options.Databse

	// SSL Mode
	if sslMode := c.options.SslMode; sslMode != "" {
		connStr += fmt.Sprintf("?sslmode=%s", sslMode)
	}

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

func (c *client) Disconnect() error {
	return nil
}

func (c *client) Migrate() error {
	s := bindata.Resource(migrations.AssetNames(),
		func(name string) ([]byte, error) {
			return migrations.Asset(name)
		})
	d, err := bindata.WithInstance(s)
	if err != nil {
		return fmt.Errorf("Could not create migrations reader: %v", err)
	}

	driver, err := postgresMigrate.WithInstance(c.db, &postgresMigrate.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("go-binddata", d, "postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			c.log.Debug().Msg("Database already up to date")
		} else {
			return err
		}
	}
	return nil
}

func (c *client) Select(query string, args ...any) *sql.Row {
	return c.db.QueryRow(query, args...)
}

func (c *client) Execute(query string, args ...any) error {
	exec, err := c.db.Exec(query, args...)
	if err != nil {
		return err
	}
	affected, _ := exec.RowsAffected()
	c.log.Debug().Int64("affected", affected).Str("query", query).Msg("Query executed")
	return nil
}
