package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	_ "github.com/Tanya1515/gophermarket/cmd/additional"
	storage "github.com/Tanya1515/gophermarket/cmd/storage"
	psql "github.com/Tanya1515/gophermarket/cmd/storage/postgresql"
)

type Gophermarket struct {
	storage storage.StorageInterface
	logger  zap.SugaredLogger
	secretKey string
}

func init() {
	marketAddressFlag = flag.String("a", "localhost:8080", "gophermarket address")
	storageUrlFlag = flag.String("d", "localhost:5432", "database url")
	// accrualSystemAddressFlag = flag.String("r", "localhost:8081", "acccrual system address")
}

var (
	marketAddressFlag *string
	storageUrlFlag    *string
	// accrualSystemAddressFlag *string
)

func main() {
	var GM Gophermarket
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	MarketLogger := *logger.Sugar()

	GM.logger = MarketLogger

	defer GM.logger.Sync()

	flag.Parse()

	marketAddress, ok := os.LookupEnv("RUN_ADDRESS")
	if !ok {
		marketAddress = *marketAddressFlag
	}

	storageAddress, ok := os.LookupEnv("DATABASE_URI")
	if !ok {
		storageAddress = *storageUrlFlag
	}

	storageAddressArgs := strings.Split(storageAddress, ":")

	// accrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	// if !ok {
	// 	accrualSystemAddress = *accrualSystemAddressFlag
	// }

	GM.storage = &psql.PostgreSQL{Address: storageAddressArgs[0], Port: storageAddressArgs[1], UserName: "collector", Password: "postgres", DBName: "gophermarket"}
	err = GM.storage.Init()
	if err != nil {
		panic(err)
	}
	GM.logger.Infow(
		"Gophermarket starts working",
		"address: ", marketAddress,
	)

	GM.secretKey = "secretKey"

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", GM.RegisterUser())
		r.Post("/api/user/login", GM.AuthentificateUser())
		r.Post("/api/user/orders", GM.MiddlewareCheckUser(GM.AddOrdersInfobyUser()))
		r.Get("/api/user/orders", GM.MiddlewareCheckUser(GM.GetOrdersInfobyUser()))
		r.Get("/api/user/balance", GM.MiddlewareCheckUser(GM.GetUserBalance()))
		r.Get("/api/user/withdrawals", GM.MiddlewareCheckUser(GM.GetUserWastes()))
		r.Post("/api/user/balance/withdraw", GM.MiddlewareCheckUser(GM.PayByPoints()))

	})

	err = http.ListenAndServe(marketAddress, r)
	if err != nil {
		GM.logger.Fatalw(err.Error(), "event", "start server")
	}
}
