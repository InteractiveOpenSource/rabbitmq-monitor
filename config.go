package main

import "fmt"

type ServerConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Vhost string
}

func (c *ServerConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("invalid host, host=%s -> --host [host]", c.Host)
	}

	if c.User == "" {
		return fmt.Errorf("user is required, user=%s -> --user [user]", c.User)
	}

	if c.Password == "" {
		return fmt.Errorf("password is required, --password [password]")
	}

	if c.Vhost == "" {
		return fmt.Errorf("vhost is required, --vhost [vhost]")
	}

	return nil
}
