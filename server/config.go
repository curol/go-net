package server

import (
	"reflect"
	"time"
)

const (
	MB = 1024 * 1024 // 1 MB
)

type Config struct {
	Network         string
	Address         string
	MaxConnections  int
	MaxReadSize     int
	MaxResponseSize int
	DeadLine        time.Time
}

func NewConfig(options *Config) Config {
	// Defaults
	defaultConfig := Config{
		Network:         "tcp",
		Address:         ":8080",
		MaxReadSize:     5 * 1024 * 1024,                  // 5 * 1MB
		MaxResponseSize: 5 * 1024 * 1024,                  // 5 * 1MB
		DeadLine:        time.Now().Add(10 * time.Minute), // 10 seconds
	}

	// Return default config if options is nil
	if options == nil {
		return defaultConfig
	}

	// Return merged config
	return defaultConfig.merge(*options)
}

// Returns new Config with values from config merged into c.
func (c *Config) merge(config Config) Config {
	return mergeConfigs(*c, config)
}

// mergeConfigs takes two Config structs as arguments.
// It uses reflection to iterate over the fields of the structs.
// If the field in the second struct is not zero, it overwrites the corresponding field in the first struct.
// The function then returns the first struct, which now contains the merged values.
func mergeConfigs(a, b Config) Config {
	va := reflect.ValueOf(&a).Elem()
	vb := reflect.ValueOf(&b).Elem()

	for i := 0; i < va.NumField(); i++ {
		vaField := va.Field(i)
		vbField := vb.Field(i)

		if vbField.Interface() != reflect.Zero(vbField.Type()).Interface() {
			vaField.Set(vbField)
		}
	}

	return a
}
