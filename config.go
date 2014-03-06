package redeo

// Server configuration
type Config struct {
	Proto string
	Addr  string
}

// Default configuration is used when nil is passed to NewServer
var DefaultConfig = &Config{
	Proto: "tcp",
	Addr:  "127.0.0.1:9736",
}
