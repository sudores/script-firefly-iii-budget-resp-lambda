package cnf

import "github.com/caarlos0/env"

// Cnf the config object with configuration parameters
type Cnf struct {
	LogLevel           string            `env:"LOG_LEVEL" envDefault:"debug"`
	ListenAddr         string            `env:"LISTEN_ADDRESS" envDefault:":3000"`
	BudgetPathRelation map[string]string `env:"BUDGET_PATH_RELATION"`
	FFIToken           string            `env:"FFI_TOKEN,required"`
	FFIURL             string            `env:"FFI_URL,required"`
}

// Parse parses the env variables defined in Cnf tags to Cnf struct pointer
func Parse() (*Cnf, error) {
	cnf := Cnf{}
	if err := env.Parse(&cnf); err != nil {
		return nil, err
	}
	return &cnf, nil
}
