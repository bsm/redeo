package redeo

// Server configuration
type Config struct {
	Addr   string
	Socket string
}

// Default configuration is used when nil is passed to NewServer
var DefaultConfig = &Config{
	Addr: "127.0.0.1:9736",
}
