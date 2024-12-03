package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	interval  = 7000
	noTimeout = -1
)

var topic1 = "orders"
var topic2 = "get-by-id"
var topic3 = "get-order"

func getOrderByUID(db *sql.DB, orderUID string) (*Order, error) {

	err := incrementOrderFrequency(db, orderUID, 1)
	if err != nil {
		log.Printf("Can not increment frequency")
	}

	frequency, err := getOrderFrequency(db, orderUID)
	if err != nil {
		log.Printf("Can not get frequency")
	} else if frequency > 4 {
		log.Printf("Getting order from cache")
		return storeOrders[orderUID], nil
	}

	// Инициализируем основной объект Order
	order := &Order{}

	// Получение данных из таблицы orders
	err = db.QueryRow(`
		SELECT order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.CustomerID,
		&order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Заказ не найден
		}
		return nil, err
	}

	// Получение данных из таблицы delivery
	err = db.QueryRow(`
		SELECT name, phone, zip, city, address, region, email
		FROM delivery WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		return nil, err
	}

	// Получение данных из таблицы payment
	err = db.QueryRow(`
		SELECT transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payment WHERE order_uid = $1
	`, orderUID).Scan(
		&order.Payment.Transaction, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount,
		&order.Payment.PaymentDT, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return nil, err
	}

	// Получение данных из таблицы items
	rows, err := db.Query(`
		SELECT chrt_id, order_uid, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`, orderUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Итерация по всем элементам заказа
	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ChrtID, &item.OrderUID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	if frequency == 4 {
		cacheSave(order)
	}

	return order, nil
}

func saveOrder(db *sql.DB, order Order) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Сохранение заказа
	_, err = tx.Exec(`
		INSERT INTO orders (order_uid, track_number, entry, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, frequency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 0);
		`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}

	// Сохранение доставки
	_, err = tx.Exec(`
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return err
	}

	// Сохранение платежей
	_, err = tx.Exec(`
		INSERT INTO payment (order_uid, transaction, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`, order.OrderUID, order.Payment.Transaction, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return err
	}

	// Сохранение элементов
	for _, item := range order.Items {
		_, err = tx.Exec(`
			INSERT INTO items (chrt_id, order_uid, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
		`, item.ChrtID, order.OrderUID, item.TrackNumber, item.Price, item.RID, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (c *Consumer) Stop() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		return err
	}

	log.Println("Committed offset")
	return c.consumer.Close()
}

func incrementOrderFrequency(db *sql.DB, orderUID string, increment int) error {
	_, err := db.Exec(`
		UPDATE orders
		SET frequency = frequency + $1
		WHERE order_uid = $2
	`, increment, orderUID)
	if err != nil {
		return fmt.Errorf("failed to increment frequency for order %s: %w", orderUID, err)
	}

	log.Printf("Successfully incremented frequency by %d for order %s", increment, orderUID)
	return nil
}

func getOrderFrequency(db *sql.DB, orderUID string) (int, error) {
	var frequency int

	// Выполняем SQL-запрос для получения значения frequency
	err := db.QueryRow(`
		SELECT frequency
		FROM orders
		WHERE order_uid = $1
	`, orderUID).Scan(&frequency)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("order with UID %s not found", orderUID)
		}
		return 0, fmt.Errorf("failed to get frequency for order %s: %w", orderUID, err)
	}

	return frequency, nil
}

func main() {
	// Подключение к базе данных
	connectionString := "user=ibushkovskij password=psg dbname=orders host=localhost port=54358 sslmode=disable"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %s", err)
	}
	defer db.Close()

	cacheStarting(db)

	consumer, err := NewConsumer(topic1)
	if err != nil {
		log.Fatalf("Cannot create new Consumer: %s", err)
	}

	go func() {
		consumer.Start(db)
	}()

	cons2, err := NewConsumer(topic2)
	if err != nil {
		log.Fatalf("Cannot create new Consumer: %s", err)
	}

	prod, err := NewProd()

	go func() {
		cons2.StartToSelect(prod, db)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Fatal(consumer.Stop())
}
