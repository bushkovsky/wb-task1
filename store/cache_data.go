package main

import (
	"database/sql"
	"fmt"
	"log"
)

var storeOrders = make(map[string]*Order)

func cacheSave(order *Order) {
	storeOrders[order.OrderUID] = order
	log.Printf("Success saved")
}

func cacheStarting(db *sql.DB) {
	orders, err := fetchOrdersWithFrequency(db, 4)
	if err != nil {
		log.Println("Can not cache most using data")
	}

	for _, ord := range orders {
		storeOrders[ord.OrderUID] = ord
	}

}

func fetchOrdersWithFrequency(db *sql.DB, minFrequency int) ([]*Order, error) {
	rows, err := db.Query(`
		SELECT 
			o.order_uid, o.track_number, o.entry, o.locale, o.customer_id, 
			o.delivery_service, o.shardkey, o.sm_id, o.date_created, 
			o.oof_shard,
			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			p.transaction, p.currency, p.provider, p.amount, p.payment_dt, 
			p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
			i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, 
			i.size, i.total_price, i.nm_id, i.brand, i.status
		FROM orders o
		LEFT JOIN delivery d ON o.order_uid = d.order_uid
		LEFT JOIN payment p ON o.order_uid = p.order_uid
		LEFT JOIN items i ON o.order_uid = i.order_uid
		WHERE o.frequency > $1;
	`, minFrequency)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order

	for rows.Next() {
		var order Order
		var item Item
		var delivery Delivery
		var payment Payment

		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.CustomerID,
			&order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated,
			&order.OofShard,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address,
			&delivery.Region, &delivery.Email,
			&payment.Transaction, &payment.Currency, &payment.Provider, &payment.Amount,
			&payment.PaymentDT, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal,
			&payment.CustomFee,
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Заполняем данные заказа
		order.Delivery = delivery
		order.Payment = payment
		order.Items = append(order.Items, item)

		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return orders, nil
}
