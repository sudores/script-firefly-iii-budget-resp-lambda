package cnf

import "github.com/caarlos0/env/v9"

// Cnf the config object with configuration parameters
type Cnf struct {
	LogLevel           string         `env:"LOG_LEVEL" envDefault:"info"`
	ListenAddr         string         `env:"LISTEN_ADDRESS" envDefault:":3000"`
	FFIToken           string         `env:"FFI_TOKEN,required"`
	FFIURL             string         `env:"FFI_URL,required"`
	BudgetPathRelation map[string]int `env:"BUDGET_PATH_RELATION"`
}

// Parse parses the env variables defined in Cnf tags to Cnf struct pointer
func Parse() (*Cnf, error) {
	cnf := Cnf{}
	if err := env.Parse(&cnf); err != nil {
		return nil, err
	}
	return &cnf, nil
}
