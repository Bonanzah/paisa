package server

import (
	"strings"
	"time"
	"unicode"

	"github.com/ananthakumaran/paisa/internal/model/receipt_item"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ReceiptItemRequest struct {
	Name     string          `json:"name" binding:"required"`
	Brand    string          `json:"brand"`
	Variant  string          `json:"variant"`
	Unit     string          `json:"unit" binding:"required"`
	Quantity decimal.Decimal `json:"quantity" binding:"required"`
	Price    decimal.Decimal `json:"price" binding:"required"`
}

type ReceiptRequest struct {
	Store string               `json:"store" binding:"required"`
	Date  string               `json:"date" binding:"required"`
	Items []ReceiptItemRequest `json:"items" binding:"required,min=1"`
}

func PostReceipt(db *gorm.DB, request ReceiptRequest) gin.H {
	date, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		return gin.H{"success": false, "error": "invalid date format, expected YYYY-MM-DD"}
	}

	var items []receipt_item.ReceiptItem
	for _, ri := range request.Items {
		if ri.Quantity.IsZero() || ri.Quantity.IsNegative() {
			return gin.H{"success": false, "error": "quantity must be positive for item: " + ri.Name}
		}

		unitPrice := ri.Price.Div(ri.Quantity)

		items = append(items, receipt_item.ReceiptItem{
			Date:      date,
			Store:     normalizeToTitle(request.Store),
			Name:      normalizeToLower(ri.Name),
			Brand:     normalizeToTitle(ri.Brand),
			Variant:   normalizeToLower(ri.Variant),
			Unit:      ri.Unit,
			Quantity:  ri.Quantity,
			Price:     ri.Price,
			UnitPrice: unitPrice,
		})
	}

	err = receipt_item.Create(db, items)
	if err != nil {
		return gin.H{"success": false, "error": err.Error()}
	}

	return gin.H{"success": true, "count": len(items)}
}

type PriceMover struct {
	Name        string          `json:"name"`
	Unit        string          `json:"unit"`
	Change      decimal.Decimal `json:"change"`
	LatestPrice decimal.Decimal `json:"latest_price"`
}

type StoreRanking struct {
	Store string `json:"store"`
	Count int    `json:"count"`
}

type RecentEntry struct {
	Date  time.Time `json:"date"`
	Store string    `json:"store"`
	Count int       `json:"count"`
}

func GetPriceTracking(db *gorm.DB) gin.H {
	items := receipt_item.All(db)
	names := receipt_item.DistinctNames(db)
	stores := receipt_item.DistinctStores(db)

	if len(items) == 0 {
		return gin.H{
			"items_count":   0,
			"stores_count":  0,
			"avg_change":    decimal.Zero,
			"increases":     []PriceMover{},
			"decreases":     []PriceMover{},
			"store_ranking": []StoreRanking{},
			"recent":        []RecentEntry{},
		}
	}

	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	byName := make(map[string][]receipt_item.ReceiptItem)
	for _, item := range items {
		byName[item.Name] = append(byName[item.Name], item)
	}

	var increases, decreases []PriceMover
	changeSum := decimal.Zero
	changeCount := 0

	for _, name := range names {
		nameItems := byName[name]
		if len(nameItems) < 2 {
			continue
		}

		latest := nameItems[0]
		var baseline *receipt_item.ReceiptItem
		for i := range nameItems {
			if nameItems[i].Date.Before(thirtyDaysAgo) {
				baseline = &nameItems[i]
				break
			}
		}
		if baseline == nil || baseline.UnitPrice.IsZero() {
			continue
		}

		change := latest.UnitPrice.Sub(baseline.UnitPrice).Div(baseline.UnitPrice).Mul(decimal.NewFromInt(100))
		changeSum = changeSum.Add(change)
		changeCount++

		mover := PriceMover{
			Name:        name,
			Unit:        latest.Unit,
			Change:      change,
			LatestPrice: latest.UnitPrice,
		}

		if change.IsPositive() {
			increases = append(increases, mover)
		} else if change.IsNegative() {
			decreases = append(decreases, mover)
		}
	}

	sortMovers(increases, false)
	sortMovers(decreases, true)

	if len(increases) > 5 {
		increases = increases[:5]
	}
	if len(decreases) > 5 {
		decreases = decreases[:5]
	}

	avgChange := decimal.Zero
	if changeCount > 0 {
		avgChange = changeSum.Div(decimal.NewFromInt(int64(changeCount)))
	}

	storeWins := make(map[string]int)
	for _, name := range names {
		nameItems := byName[name]
		latestByStore := make(map[string]decimal.Decimal)
		for _, item := range nameItems {
			if _, exists := latestByStore[item.Store]; !exists {
				latestByStore[item.Store] = item.UnitPrice
			}
		}
		cheapestStore := ""
		cheapestPrice := decimal.Zero
		for store, price := range latestByStore {
			if cheapestStore == "" || price.LessThan(cheapestPrice) {
				cheapestStore = store
				cheapestPrice = price
			}
		}
		if cheapestStore != "" {
			storeWins[cheapestStore]++
		}
	}

	var storeRanking []StoreRanking
	for store, count := range storeWins {
		storeRanking = append(storeRanking, StoreRanking{Store: store, Count: count})
	}
	sortStoreRanking(storeRanking)

	type dateStore struct {
		Date  time.Time
		Store string
	}
	recentMap := make(map[dateStore]int)
	for _, item := range items {
		key := dateStore{Date: item.Date, Store: item.Store}
		recentMap[key]++
	}
	var recent []RecentEntry
	for key, count := range recentMap {
		recent = append(recent, RecentEntry{Date: key.Date, Store: key.Store, Count: count})
	}
	sortRecentEntries(recent)
	if len(recent) > 10 {
		recent = recent[:10]
	}

	return gin.H{
		"items_count":   len(names),
		"stores_count":  len(stores),
		"avg_change":    avgChange,
		"increases":     increases,
		"decreases":     decreases,
		"store_ranking": storeRanking,
		"recent":        recent,
	}
}

func GetPriceTrackingItems(db *gorm.DB) gin.H {
	return gin.H{"items": receipt_item.DistinctNames(db)}
}

func GetPriceTrackingItem(db *gorm.DB, name string) gin.H {
	items := receipt_item.ByName(db, name)

	if len(items) == 0 {
		return gin.H{
			"name":     name,
			"entries":  []receipt_item.ReceiptItem{},
			"stores":   []string{},
			"brands":   []string{},
			"variants": []string{},
			"latest":   nil,
			"avg":      decimal.Zero,
			"change":   decimal.Zero,
		}
	}

	storeSet := make(map[string]bool)
	brandSet := make(map[string]bool)
	variantSet := make(map[string]bool)
	sum := decimal.Zero

	for _, item := range items {
		storeSet[item.Store] = true
		if item.Brand != "" {
			brandSet[item.Brand] = true
		}
		if item.Variant != "" {
			variantSet[item.Variant] = true
		}
		sum = sum.Add(item.UnitPrice)
	}

	avg := sum.Div(decimal.NewFromInt(int64(len(items))))
	latest := items[0]

	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	change := decimal.Zero
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].Date.Before(ninetyDaysAgo) && !items[i].UnitPrice.IsZero() {
			change = latest.UnitPrice.Sub(items[i].UnitPrice).Div(items[i].UnitPrice).Mul(decimal.NewFromInt(100))
			break
		}
	}

	return gin.H{
		"name":     name,
		"entries":  items,
		"stores":   setToSlice(storeSet),
		"brands":   setToSlice(brandSet),
		"variants": setToSlice(variantSet),
		"latest":   latest,
		"avg":      avg,
		"change":   change,
	}
}

type UpdateReceiptItemRequest struct {
	Store    *string          `json:"store"`
	Brand    *string          `json:"brand"`
	Variant  *string          `json:"variant"`
	Quantity *decimal.Decimal `json:"quantity"`
	Price    *decimal.Decimal `json:"price"`
}

func UpdateReceiptItem(db *gorm.DB, id uint, request UpdateReceiptItemRequest) (gin.H, int) {
	item, err := receipt_item.FindByID(db, id)
	if err != nil {
		return gin.H{"success": false, "error": "item not found"}, 404
	}

	if request.Store != nil {
		item.Store = normalizeToTitle(*request.Store)
	}
	if request.Brand != nil {
		item.Brand = normalizeToTitle(*request.Brand)
	}
	if request.Variant != nil {
		item.Variant = normalizeToLower(*request.Variant)
	}

	if request.Quantity != nil {
		if request.Quantity.IsZero() || request.Quantity.IsNegative() {
			return gin.H{"success": false, "error": "quantity must be positive"}, 400
		}
		item.Quantity = *request.Quantity
	}
	if request.Price != nil {
		if request.Price.IsNegative() {
			return gin.H{"success": false, "error": "price must not be negative"}, 400
		}
		item.Price = *request.Price
	}

	if request.Quantity != nil || request.Price != nil {
		item.UnitPrice = item.Price.Div(item.Quantity)
	}

	err = receipt_item.Update(db, item)
	if err != nil {
		return gin.H{"success": false, "error": err.Error()}, 500
	}

	return gin.H{"success": true, "item": item}, 200
}

func DeleteReceiptItem(db *gorm.DB, id uint) (gin.H, int) {
	_, err := receipt_item.FindByID(db, id)
	if err != nil {
		return gin.H{"success": false, "error": "item not found"}, 404
	}

	err = receipt_item.Delete(db, id)
	if err != nil {
		return gin.H{"success": false, "error": err.Error()}, 500
	}

	return gin.H{"success": true}, 200
}

func sortMovers(movers []PriceMover, ascending bool) {
	for i := 0; i < len(movers); i++ {
		for j := i + 1; j < len(movers); j++ {
			if ascending {
				if movers[j].Change.LessThan(movers[i].Change) {
					movers[i], movers[j] = movers[j], movers[i]
				}
			} else {
				if movers[j].Change.GreaterThan(movers[i].Change) {
					movers[i], movers[j] = movers[j], movers[i]
				}
			}
		}
	}
}

func sortStoreRanking(rankings []StoreRanking) {
	for i := 0; i < len(rankings); i++ {
		for j := i + 1; j < len(rankings); j++ {
			if rankings[j].Count > rankings[i].Count {
				rankings[i], rankings[j] = rankings[j], rankings[i]
			}
		}
	}
}

func sortRecentEntries(entries []RecentEntry) {
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Date.After(entries[i].Date) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

func setToSlice(set map[string]bool) []string {
	var result []string
	for key := range set {
		result = append(result, key)
	}
	return result
}

func normalizeToLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeToTitle(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	words := strings.Fields(s)
	for i, word := range words {
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
