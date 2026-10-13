package config

import (
	"log/slog"
	"sync"
	"time"

	"go.rtnl.ai/confire"
	"go.rtnl.ai/confire/validate"
	"go.rtnl.ai/x/rlog"
)

const Prefix = "wilson"

type Config struct {
	LogLevel       rlog.LevelDecoder `split_words:"true" default:"info" desc:"specify the verbosity of logging (trace, debug, info, warn, error, fatal, or panic)"`
	ConsoleLog     bool              `split_words:"true" default:"false" desc:"if true logs human readable text output instead of json"`
	BindAddr       string            `split_words:"true" default:":25" desc:"the address to bind the smtp server to (e.g. :25, :587, :465)"`
	Domain         string            `split_words:"true" default:"localhost" desc:"the domain name of the server and recipient email addresses"`
	ReadTimeout    time.Duration     `split_words:"true" default:"4s" desc:"the read timeout for the smtp server"`
	WriteTimeout   time.Duration     `split_words:"true" default:"16s" desc:"the write timeout for the smtp server"`
	MaxMessageSize int64             `split_words:"true" default:"10485760" desc:"the maximum size of a message to accept (in bytes)"`
	MaxRecipients  int               `split_words:"true" default:"32" desc:"the maximum number of recipients to accept"`
	UseTLS         bool              `split_words:"true" default:"false" desc:"if true the server will use TLS"`
	TLSCerts       string            `split_words:"true" default:"" desc:"the path to the TLS certificate file"`
	TLSKey         string            `split_words:"true" default:"" desc:"the path to the TLS key file"`
}

func New() (conf *Config, err error) {
	// NOTE: confire.Process calls Validate() internally.
	conf = &Config{}
	if err = confire.Process(Prefix, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func (c Config) Validate() (err error) {
	if c.UseTLS {
		if c.TLSCerts == "" {
			err = confire.Join(err, confire.Required("", "tls_certs"))
		}
		if c.TLSKey == "" {
			err = confire.Join(err, confire.Required("", "tls_key"))
		}
	}
	return err
}

// GetLogLevel returns the log level for the config.
func (c Config) GetLogLevel() slog.Level {
	return c.LogLevel.Level()
}

//============================================================================
// Config Package Management
//============================================================================

var (
	mu   sync.RWMutex
	err  error // only written by New() inside Get's sync.Once, and cleared on successful Set
	load sync.Once
	conf *Config
)

func Get() (Config, error) {
	load.Do(func() {
		mu.Lock()
		defer mu.Unlock()

		if conf == nil {
			conf, err = New()
		}
	})
	mu.RLock()
	defer mu.RUnlock()
	if conf != nil {
		return *conf, err
	}
	return Config{}, err
}

func MustGet() Config {
	conf, err := Get()
	if err != nil {
		panic(err)
	}
	return conf
}

func Set(c Config) error {
	mu.Lock()
	defer mu.Unlock()

	if err := validate.Validate(c); err != nil {
		return err
	}

	conf = &c
	err = nil

	return nil
}

func Reset() {
	mu.Lock()
	defer mu.Unlock()
	conf = nil
	err = nil
	load = sync.Once{}
}
