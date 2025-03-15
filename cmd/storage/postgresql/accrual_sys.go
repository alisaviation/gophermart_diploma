package storage

import(
	add "github.com/Tanya1515/gophermarket/cmd/additional" 
)

func (db *PostgreSQL) FinishUserOrder(order add.Order) (err error) {
	// проверяем статус заказа, если INVALID - записываем в базу

	// если статус заказа - PROCESSED и есть accrual: 
	// join по id пользователя, выбирая заказ по его номеру и увеличиваем сумму баллов на балансе
	// пользователя 
	
	// сохраняем статус заказа

	// убираем заказ из мапы
	return 
}

func (db *PostgreSQL) StartProcessingUserOrder(orders *[]add.Order) (err error) {

	// вытаскиваем из базы данных новые заказы (со статусом NEW)
	// меняем статус всех заказов на processing

	// возвращаем новые заказы
	
	return
}