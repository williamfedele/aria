package config

type Config struct {
	LibraryDir string
}

// TODO: store this in a config file. have some exec flag to set it
func NewConfig() Config {
	return Config{LibraryDir: "/Users/will/Music/library"}
}
