package main

import "time"

type Part struct {
	Cve   string  `json:"cve" bson:"cve"`
	Name  string  `json:"name" bson:"name"`
	Price float64 `json:"price" bson:"price"`
}

type Catalog struct {
	Start_date  time.Time `json:"start_date" bson:"start_date"`
	End_date    time.Time `json:"end_date" bson:"end_date"`
	Brand       string    `json:"brand" bson:"brand"`
	Parts       []Part    `json:"parts" bson:"parts"`
	Parts_Count int       `json:"parts_count" bson:"parts_count"`
}

type Stats struct {
	PartsUpdate      PartsUpdate `bson:"parts_update" json:"parts_update"`
	ObsoletesUpdate  time.Time   `bson:"obsoletes_update" json:"obsoletes_update"`
	InventoryUpdate  time.Time   `bson:"inventory_update" json:"inventory_update"`
	BackordersUpdate time.Time   `bson:"backorders_update" json:"backorders_update"`
	WelcomeMessage   string      `bson:"welcome_message" json:"welcome_message"`
}

type PartsUpdate struct {
	Honda CarUpdate `bson:"Honda" json:"Honda"`
	Acura CarUpdate `bson:"Acura" json:"Acura"`
}

type CarUpdate struct {
	StartDate time.Time `bson:"start_date" json:"start_date"`
	EndDate   time.Time `bson:"end_date" json:"end_date"`
}

type InventoryParts struct {
	Cve  string
	Name string
	Qty  int
}

type Inventory struct {
	NoDealer     int              `json:"no_dealer" bson:"no_dealer"`
	Name         string           `json:"name" bson:"name"`
	ContactName  string           `json:"contact_name" bson:"contact_name"`
	ContactEmail string           `json:"contact_email" bson:"contact_email"`
	ContactPhone string           `json:"contact_phone" bson:"contact_phone"`
	Parts        []InventoryParts `json:"parts" bson:"parts"`
}

type ObsoletesParts struct {
	Cve  string
	Name string
	Qty  int
}

type Obsoletes struct {
	NoDealer     int              `json:"no_dealer" bson:"no_dealer"`
	Name         string           `json:"name" bson:"name"`
	ContactName  string           `json:"contact_name" bson:"contact_name"`
	ContactEmail string           `json:"contact_email" bson:"contact_email"`
	ContactPhone string           `json:"contact_phone" bson:"contact_phone"`
	Parts        []ObsoletesParts `json:"parts" bson:"parts"`
}

type BackordersParts struct {
	Cve       string    `json:"cve" bson:"cve"`
	Cve_Alt   string    `json:"cve_alt" bson:"cve_alt"`
	NoDealer  int       `json:"no_dealer" bson:"no_dealer"`
	NoOrder   string    `json:"no_order" bson:"no_order"`
	NoPedido  string    `json:"no_pedido" bson:"no_pedido"`
	Qty       int       `json:"qty" bson:"qty"`
	OrderDate time.Time `json:"order_date" bson:"order_date"`
	Forecast  string    `json:"forecast" bson:"forecast"`
}
