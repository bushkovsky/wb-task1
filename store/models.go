package main

// Основная структура для таблицы orders

type OrderID struct {
	OrderUID string `json:"order_uid" gorm:"primaryKey"`
}

type Order struct {
	OrderUID        string   `json:"order_uid" gorm:"primaryKey"`
	TrackNumber     string   `json:"track_number"`
	Entry           string   `json:"entry"`
	Locale          string   `json:"locale"`
	CustomerID      string   `json:"customer_id"`
	DeliveryService string   `json:"delivery_service"`
	ShardKey        string   `json:"shardkey"`
	SmID            int      `json:"sm_id"`
	DateCreated     string   `json:"date_created"`
	OofShard        string   `json:"oof_shard"`
	Delivery        Delivery `json:"delivery" gorm:"foreignKey:OrderUID"`
	Payment         Payment  `json:"payment" gorm:"foreignKey:OrderUID"`
	Items           []Item   `json:"items" gorm:"foreignKey:OrderUID"`
}

// Структура для таблицы delivery
type Delivery struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderUID string `json:"order_uid" gorm:"index;not null"` // Индекс, чтобы связывать с Order
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Zip      string `json:"zip"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Region   string `json:"region"`
	Email    string `json:"email"`
}

// Структура для таблицы payment
type Payment struct {
	ID           uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderUID     string `json:"order_uid" gorm:"index;not null"` // Индекс, чтобы связывать с Order
	Transaction  string `json:"transaction"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

// Структура для таблицы items
type Item struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	ChrtID      int64  `json:"chrt_id"`
	OrderUID    string `json:"order_uid" gorm:"index;not null"` // Индекс, чтобы связывать с Order
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int64  `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}
