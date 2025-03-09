package storage

import(
	add "github.com/Tanya1515/gophermarket/cmd/additional" 
)

func (db *PostgreSQL) FinishUserOrder(order add.Order) (err error) {
	// проверяем статус заказа, если INVALID - записываем в базу

	// если статус заказа - PROCESSED: 
	// join по id пользователя, выбирая заказ по его номеру и увеличиваем сумму баллов на балансе
	// пользователя 
	
	// сохраняем статус заказа
	return 
}

func (db *PostgreSQL) StartProcessingUserOrder(orders *[]add.Order) (err error) {

	// вытаскиваем из базы данных новые заказы 
	// сохраняем все заказы поочередно в срез заказов
	// для каждого сохраненного заказа меняем его статус в PROCESSING 
	
	return
}