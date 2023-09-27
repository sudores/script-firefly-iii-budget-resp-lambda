package main

import (
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sudores/script-firefly-iii-budget-resp/cnf"
	fireflyiii "github.com/sudores/script-firefly-iii-budget-resp/firefly-iii"
)

func main() {
	cfg, err := cnf.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	}

	loggingInit(cfg.LogLevel)
	log.Info().Msg("Logging setup success")

	ffi := fireflyiii.NewFireflyiiiConnection(cfg.FFIToken, cfg.FFIURL, cfg.BudgetPathRelation)

	rtr := http.NewServeMux()
	rtr.Handle("/", ffi)

	log.Info().Msgf("Starting serving. Listening at %s", cfg.ListenAddr)
	http.Handle("/", rtr)
	if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

// loggingInit setups the logging of whole app
func loggingInit(logLevel string) {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Fatal().Msgf(`Log level "%s" is unrecognized. Eligible log levels are: trace, debug, info, err, fatal, panic`, logLevel)
	}
	zerolog.SetGlobalLevel(level)
	log.Debug().Msg("Logger initialized")
}
