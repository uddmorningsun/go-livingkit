package config

// Connection containing the settings for all databases and caches configurations, design inspired by Django.
type Connection struct {
	Backend  string                 `yaml:"backend" json:"backend"`
	Address  string                 `yaml:"address" json:"address"`
	User     string                 `yaml:"user" json:"user"`
	Password string                 `yaml:"password" json:"password"`
	Name     string                 `yaml:"name" json:"name"`
	Enabled  bool                   `yaml:"enabled" json:"enabled"`
	Options  map[string]interface{} `yaml:"options" json:"options"`
}
