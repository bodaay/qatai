package db

type QataiDatabase interface {
	GetConfig(key string) (*Config, error)
	SetConfig(config *Config) error
	GetUser(id string) (*User, error)
	SetUser(user *User) error
}
