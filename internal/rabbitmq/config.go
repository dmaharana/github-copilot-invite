package rabbitmq

// Config holds RabbitMQ configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	VHost    string
}

// NewConfig creates a new RabbitMQ configuration
func NewConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     5672,
		Username: "guest",
		Password: "guest",
		VHost:    "/",
	}
}
