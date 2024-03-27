package ethereum

const (
	defaultHostname = "localhost"
	defaultPort     = 4444
)

type Config struct {
	Hostname string
	Port     int
	Username string
	Password string
}

func NewConfig() *Config {
	return &Config{
		Hostname: defaultHostname,
		Port:     defaultPort,
	}
}
