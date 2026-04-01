package receipt_item

import (
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReceiptItem struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Date      time.Time       `json:"date"`
	Store     string          `json:"store"`
	Name      string          `json:"name"`
	Brand     string          `json:"brand"`
	Variant   string          `json:"variant"`
	Unit      string          `json:"unit"`
	Quantity  decimal.Decimal `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
	UnitPrice decimal.Decimal `json:"unit_price"`
}

func Create(db *gorm.DB, items []ReceiptItem) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			err := tx.Create(&item).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func All(db *gorm.DB) []ReceiptItem {
	var items []ReceiptItem
	result := db.Order("date DESC").Find(&items)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return items
}

func ByName(db *gorm.DB, name string) []ReceiptItem {
	var items []ReceiptItem
	result := db.Where("name = ?", name).Order("date DESC").Find(&items)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return items
}

func DistinctNames(db *gorm.DB) []string {
	var names []string
	result := db.Model(&ReceiptItem{}).Distinct("name").Order("name").Pluck("name", &names)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return names
}

func DistinctStores(db *gorm.DB) []string {
	var stores []string
	result := db.Model(&ReceiptItem{}).Distinct("store").Order("store").Pluck("store", &stores)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return stores
}
