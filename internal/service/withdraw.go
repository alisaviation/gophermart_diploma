package service

import (
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"gophermart/internal/interfaces"
	"gophermart/internal/repository"
)

type WithdrawService struct {
	WithdrawRepository *repository.WithdrawRepository
}

func (ws *WithdrawService) Withdraw(userID int, orderNumber string, sum decimal.Decimal) (int, error) {
	wr := ws.WithdrawRepository
	tx, err := wr.DBStorage.Conn.BeginTx(wr.DBStorage.Ctx, pgx.TxOptions{})
	if err != nil {
		return repository.NotEnoughFound, fmt.Errorf("ошибка при начале транзакции: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(wr.DBStorage.Ctx)
		}
	}()

	userBalance, err := wr.GetCurrentBalance(userID)
	if err != nil {
		_ = tx.Rollback(wr.DBStorage.Ctx)
		return repository.WithdrawTransactionError, err
	}

	if userBalance.LessThan(sum) {
		_ = tx.Rollback(wr.DBStorage.Ctx)
		return repository.NotEnoughFound, nil
	}

	if err := wr.UpdateUserBalance(userID, sum); err != nil {
		_ = tx.Rollback(wr.DBStorage.Ctx)
		return repository.WithdrawTransactionError, err
	}

	if err := wr.SaveWithdrawal(userID, orderNumber, sum); err != nil {
		_ = tx.Rollback(wr.DBStorage.Ctx)
		return repository.WithdrawTransactionError, err
	}

	if err := tx.Commit(wr.DBStorage.Ctx); err != nil {
		return repository.WithdrawTransactionError, err
	}

	return 0, nil
}

func (ws *WithdrawService) Withdrawals(userID int) ([]interfaces.WithdrawInfo, error) {
	return ws.WithdrawRepository.Withdrawals(userID)
}
