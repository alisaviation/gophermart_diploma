package securitate

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Repinoid/kurs/internal/models"
	pgx "github.com/jackc/pgx/v5"
)

type DBstruct struct {
	DB *pgx.Conn
}

var DataBase *DBstruct

var DBEndPoint = "postgres://postgres:passwordas@localhost:5432/forgo"

//var DBEndPoint = "postgres://postgres:passwordas@forgo.c7wegmiakpkw.us-west-1.rds.amazonaws.com:5432/forgo"

//var "accounts" = "accounts"
//var "orders" = "orders"
//var "tokens" = "tokens"

// соединение с базой данных
func ConnectUsersTable(ctx context.Context, DBEndPoint string) (*DBstruct, error) {
	dataBase := &DBstruct{}
	baza, err := pgx.Connect(ctx, DBEndPoint)
	if err != nil {
		return nil, fmt.Errorf("can't connect to DB %s err %w", DBEndPoint, err)
	}
	dataBase.DB = baza
	return dataBase, nil
}

func (dataBase *DBstruct) UsersTableCreation(ctx context.Context) error {

	db := dataBase.DB
	db.Exec(ctx, "CREATE EXTENSION pgcrypto;") // расширение для хэширования паролей

	creatorOrder :=
		"CREATE TABLE IF NOT EXISTS " + "accounts" +
			"(userCode INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY ," +
			"login VARCHAR(100) UNIQUE," +
			"password VARCHAR(200) NOT NULL," +
			"user_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"

	_, err := db.Exec(ctx, creatorOrder)
	if err != nil {
		return fmt.Errorf("create users table. %w", err)
	}
	return nil
}
func (dataBase *DBstruct) OrdersTableCreation(ctx context.Context) error {
	db := dataBase.DB
	creatorOrder :=
		"CREATE TABLE IF NOT EXISTS " + "orders" +
			"(id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY," +
			"userCode INT NOT NULL," +
			"orderNumber BIGINT NOT NULL UNIQUE," +
			"orderStatus VARCHAR(20)," +
			"accrual FLOAT8," +
			"uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
			"FOREIGN KEY (userCode) REFERENCES " + "accounts" + "(usercode) ON DELETE CASCADE);"

	_, err := db.Exec(ctx, creatorOrder)
	if err != nil {
		return fmt.Errorf("create orders table. %w", err)
	}
	return nil
}

func (dataBase *DBstruct) TokensTableCreation(ctx context.Context) error {
	db := dataBase.DB
	creatorOrder :=
		"CREATE TABLE IF NOT EXISTS " + "tokens" +
			"(id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY," +
			"userCode INT NOT NULL UNIQUE," +
			//			"balance FLOAT8 DEFAULT 0," +
			//			"bonus FLOAT8 DEFAULT 0," +
			"token VARCHAR(1000) NOT NULL," +
			"token_valid_until TIMESTAMP," +
			"token_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
			"FOREIGN KEY (userCode) REFERENCES " + "accounts" + "(usercode) ON DELETE CASCADE);"
	_, err := db.Exec(ctx, creatorOrder)
	if err != nil {
		return fmt.Errorf("create orders table. %w", err)
	}
	return nil
}
func (dataBase *DBstruct) WithdrawalsTableCreation(ctx context.Context) error {
	db := dataBase.DB
	creatorOrder :=
		"CREATE TABLE IF NOT EXISTS " + "withdrawn" +
			"(id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY," +
			"userCode INT NOT NULL," +
			"orderNumber BIGINT NOT NULL UNIQUE," +
			"amount FLOAT8 DEFAULT 0," +
			"processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
			"FOREIGN KEY (userCode) REFERENCES " + "accounts" + "(usercode) ON DELETE CASCADE);"
	_, err := db.Exec(ctx, creatorOrder)
	if err != nil {
		return fmt.Errorf("create orders table. %w", err)
	}
	return nil
}

func ConnectToDB(ctx context.Context) (*DBstruct, error) {
	DB, err := ConnectUsersTable(ctx, DBEndPoint)
	if err != nil {
		fmt.Printf("database connection error  %v", err)
		return nil, err
	}
	if err := DB.UsersTableCreation(ctx); err != nil {
		return nil, fmt.Errorf("UsersTableCreation %w", err)
	}
	if err := DB.OrdersTableCreation(ctx); err != nil {
		return nil, fmt.Errorf("OrdersTableCreation %w", err)
	}
	if err := DB.TokensTableCreation(ctx); err != nil {
		return nil, fmt.Errorf("TokensTableCreation %w", err)
	}
	if err := DB.WithdrawalsTableCreation(ctx); err != nil {
		return nil, fmt.Errorf("WithdrawalsTableCreation %w", err)
	}
	return DB, nil
}

func (dataBase *DBstruct) AddUser(ctx context.Context, userName, password, tokenString string) error {
	db := dataBase.DB

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error db.Begin  %[1]w", err)
	}
	defer tx.Rollback(ctx)

	order := "INSERT INTO " + "accounts" + " (login, password) VALUES ($1, crypt($2, gen_salt('md5'))) ;"
	_, err = tx.Exec(ctx, order, userName, password)
	if err != nil {
		return fmt.Errorf("add user error is %w", err)
	}
	order = "INSERT INTO tokens(userCode, token) VALUES ((select usercode from accounts where login = $1), $2) ;"
	_, err = tx.Exec(ctx, order, userName, tokenString)
	if err != nil {
		return fmt.Errorf("add TOKEN %w", err)
	}
	return tx.Commit(ctx)
}

func (dataBase *DBstruct) CheckUserPassword(ctx context.Context, userName, password string) error {
	db := dataBase.DB
	order := "SELECT (password = crypt($2, password)) AS password_match FROM " + "accounts" + " WHERE login= $1 ;"
	row := db.QueryRow(ctx, order, userName, password) // password here - what was entered
	var yes bool
	err := row.Scan(&yes)
	if err != nil {
		return fmt.Errorf("QueryRow, error is %w", err)
	}
	if !yes {
		return fmt.Errorf("password not match")
	}
	return nil
}

// nil - user exists
func (dataBase *DBstruct) IfUserExists(ctx context.Context, userName string) error {
	db := dataBase.DB
	order := "SELECT 7 from " + "accounts" + " WHERE login= $1 ;"
	row := db.QueryRow(ctx, order, userName) // password here - what was entered
	var yes int
	err := row.Scan(&yes)
	if err != nil {
		return fmt.Errorf(" QueryRow, error is %w", err)
	}
	if yes != 7 {
		return fmt.Errorf("user %s does not exist", userName)
	}
	return nil
}

func (dataBase *DBstruct) ChangePassword(ctx context.Context, userName string, password string) error {
	db := dataBase.DB
	order := "UPDATE " + "accounts" + " SET password = crypt($2, gen_salt('md5')) WHERE login= $1 ;"
	_, err := db.Exec(ctx, order, userName, password)
	if err != nil {
		return fmt.Errorf("change password error %w", err)
	}
	return nil
}

func (dataBase *DBstruct) UpdateToken(ctx context.Context, userName string, tokenString string) error {
	db := dataBase.DB
	order := "UPDATE tokens SET token = $2 WHERE userCode = (select usercode from accounts where login = $1) ;"
	_, err := db.Exec(ctx, order, userName, tokenString)
	if err != nil {
		return fmt.Errorf("add TOKEN %w", err)
	}
	return nil
}

func (dataBase *DBstruct) GetToken(ctx context.Context, userName string, tokenString *string) error {
	db := dataBase.DB
	//				получить токен из токен-таблицы  где код пользователя равен коду юзера из юзер-таблицы с именем UserName
	order := "SELECT token from " + "tokens" + " WHERE userCode = (select usercode from accounts where login = $1) ;"
	row := db.QueryRow(ctx, order, userName)
	var str string
	err := row.Scan(&str)
	if err != nil {
		return fmt.Errorf("GT %w", err)
	}
	*tokenString = str
	return nil
}

func (dataBase *DBstruct) UpLoadOrderByID(ctx context.Context, userID int64, orderNumber int64, orderStatus string, accrual float64) error {
	db := dataBase.DB
	order := "INSERT INTO orders(userCode, orderNumber, orderStatus, accrual) VALUES ($1, $2, $3, $4) ;"
	_, err := db.Exec(ctx, order, userID, orderNumber, orderStatus, accrual)
	if err != nil {
		return fmt.Errorf("add ORDER %w", err)
	}
	return nil
}

func (dataBase *DBstruct) GetIDByOrder(ctx context.Context, orderNum int64, orderID *int64) error {
	db := dataBase.DB
	order := "SELECT usercode from " + "orders" + " WHERE orderNumber =  $1 ;"
	row := db.QueryRow(ctx, order, orderNum)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		return fmt.Errorf("GT %w", err)
	}
	*orderID = id
	return nil
}

func (dataBase *DBstruct) LoginByToken(rwr http.ResponseWriter, req *http.Request) (int64, error) {

	tokenStr := req.Header.Get("Authorization")
	tokenStr, niceP := strings.CutPrefix(tokenStr, "Bearer <") // обрезаем -- Bearer <token>
	tokenStr, niceS := strings.CutSuffix(tokenStr, ">")

	var UserID int64
	if niceP && niceS {
		order := "SELECT usercode from " + "tokens" + " WHERE token =  $1 ;"
		row := dataBase.DB.QueryRow(context.Background(), order, tokenStr)
		err := row.Scan(&UserID)
		if err == nil {
			return UserID, nil
		}
	}
	rwr.WriteHeader(http.StatusUnauthorized)            // 401 — неверная пара логин/пароль;
	fmt.Fprintf(rwr, `{"status":"StatusUnauthorized"}`) // либо токена неверный формат, либо по нему нет юзера в базе
	models.Sugar.Debug("Authorization header\n")
	return 0, errors.New("Unauthorized")
}
