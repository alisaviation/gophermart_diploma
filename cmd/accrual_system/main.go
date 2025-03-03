package accrual_system

import (
	"flag"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func init() {
	marketAddressFlag = flag.String("a", "localhost:8080", "gophermarket address")
	storageUrlFlag = flag.String("d", "localhost:5432", "database url")
	accrualSystemAddressFlag = flag.String("r", "localhost:8081", "acccrual system address")
}

var (
	marketAddressFlag        *string
	storageUrlFlag           *string
	accrualSystemAddressFlag *string
)

func main() {

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	MarketLogger := *logger.Sugar()

	flag.Parse()

	marketAddress, ok := os.LookupEnv("RUN_ADDRESS")
	if !ok {
		marketAddress = *marketAddressFlag
	}

	storageAddress, ok := os.LookupEnv("DATABASE_URI")
	if !ok {
		storageAddress = *storageUrlFlag
	}

	accrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if !ok {
		accrualSystemAddress = *accrualSystemAddressFlag
	}

	MarketLogger.Infow(
		"Accural system starts working",
		"address: ", accrualSystemAddress,
	)
	defer logger.Sync()
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/api/orders/{{order_number}}")

	})
}
