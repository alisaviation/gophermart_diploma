package config

import "flag"

func ParsePlags(cfg *Config) {
	flag.StringVar(&cfg.FlagAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.FlagForDB, "d", "", "database conn link")
	flag.StringVar(&cfg.FlagAddressAS, "r", "", "address accrual system")

}
