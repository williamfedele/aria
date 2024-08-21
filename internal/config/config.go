package config

// TODO: store the config in a config file. have some exec flag to set fields like library dir

type Config struct {
	LibraryDir string
}

// TODO: store this in a config file. have some exec flag to set it
func NewConfig(libraryDir string) Config {
	return Config{LibraryDir: libraryDir}
}
