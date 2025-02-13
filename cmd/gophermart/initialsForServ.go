package main

import (
	"flag"
	"os"

	"github.com/Repinoid/kurs/internal/rual"
	"github.com/Repinoid/kurs/internal/securitate"
)

func initAgent() error {
	enva, exists := os.LookupEnv("RUN_ADDRESS")
	if exists {
		host = enva
	}
	enva, exists = os.LookupEnv("DATABASE_URI")
	if exists {
		securitate.DBEndPoint = enva
	}
	enva, exists = os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if exists {
		rual.Accrualhost = enva
	}

	var hostFlag, dbFlag, acchostFlag string
	flag.StringVar(&hostFlag, "a", host, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&dbFlag, "d", securitate.DBEndPoint, "Only -a={host:port} flag is allowed here")
	flag.StringVar(&acchostFlag, "r", rual.Accrualhost, "Only -a={host:port} flag is allowed here")
	flag.Parse()

	if _, exists := os.LookupEnv("RUN_ADDRESS"); !exists {
		host = hostFlag
	}
	if _, exists := os.LookupEnv("DATABASE_URI"); !exists {
		securitate.DBEndPoint = dbFlag
	}
	if _, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); !exists {
		rual.Accrualhost = acchostFlag
	}
	return nil
}
