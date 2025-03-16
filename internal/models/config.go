package models

type EnvVariables struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	RootUrl              string `env:"ROOT_URL"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DataBaseURL          string `env:"DATABASE_URI"`
	Secret               string `env:"SECRET_KEY"`
}
