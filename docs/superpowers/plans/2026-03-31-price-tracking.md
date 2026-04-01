# Price Tracking Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a standalone price tracking feature that accepts receipt data via API and displays item price trends over time across stores.

**Architecture:** Single `receipt_items` SQLite table via GORM, Gin handlers in `internal/server/price_tracking.go`, SvelteKit pages under `src/routes/(app)/more/price_tracking/` with D3 line charts and Tabulator tables.

**Tech Stack:** Go/Gin/GORM/decimal (backend), SvelteKit/Svelte 4/D3/Tailwind+DaisyUI+Bulma (frontend), Bun (testing)

---

### Task 1: Receipt Item Model

**Files:**
- Create: `internal/model/receipt_item/receipt_item.go`
- Modify: `internal/model/model.go`

- [ ] **Step 1: Create the receipt_item model**

Create `internal/model/receipt_item/receipt_item.go`:

```go
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
```

- [ ] **Step 2: Register model in AutoMigrate**

In `internal/model/model.go`, add import and AutoMigrate call.

Add to imports:
```go
receiptItemModel "github.com/ananthakumaran/paisa/internal/model/receipt_item"
```

Add to `AutoMigrate` function body (after the `cache.Cache` line):
```go
db.AutoMigrate(&receiptItemModel.ReceiptItem{})
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/model/receipt_item/receipt_item.go internal/model/model.go
git commit -m "feat(price-tracking): add ReceiptItem model with GORM auto-migration"
```

---

### Task 2: POST Endpoint — Receipt Ingestion

**Files:**
- Create: `internal/server/price_tracking.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Create the price_tracking handler file with POST handler**

Create `internal/server/price_tracking.go`:

```go
package server

import (
	"net/http"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/receipt_item"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ReceiptItemRequest struct {
	Name    string          `json:"name" binding:"required"`
	Brand   string          `json:"brand"`
	Variant string          `json:"variant"`
	Unit    string          `json:"unit" binding:"required"`
	Quantity decimal.Decimal `json:"quantity" binding:"required"`
	Price   decimal.Decimal `json:"price" binding:"required"`
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
			Store:     request.Store,
			Name:      ri.Name,
			Brand:     ri.Brand,
			Variant:   ri.Variant,
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
```

- [ ] **Step 2: Register POST route in server.go**

In `internal/server/server.go`, add the route before the `router.NoRoute` block:

```go
router.POST("/api/price_tracking/receipt", func(c *gin.Context) {
	var request ReceiptRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, PostReceipt(db, request))
})
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/server/price_tracking.go internal/server/server.go
git commit -m "feat(price-tracking): add POST /api/price_tracking/receipt endpoint"
```

---

### Task 3: GET Endpoints — Dashboard & Item Detail

**Files:**
- Modify: `internal/server/price_tracking.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Add GET handler functions to price_tracking.go**

Append to `internal/server/price_tracking.go`:

```go
type PriceMover struct {
	Name       string          `json:"name"`
	Unit       string          `json:"unit"`
	Change     decimal.Decimal `json:"change"`
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
			"items_count":  0,
			"stores_count": 0,
			"avg_change":   decimal.Zero,
			"increases":    []PriceMover{},
			"decreases":    []PriceMover{},
			"store_ranking": []StoreRanking{},
			"recent":       []RecentEntry{},
		}
	}

	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Group by name for price change calculations
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

		latest := nameItems[0] // already sorted DESC
		// Find the most recent item before the 30-day window
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

	// Sort increases descending, decreases ascending (most negative first)
	sortMovers(increases, false)
	sortMovers(decreases, true)

	if len(increases) > 5 {
		increases = increases[:5]
	}
	if len(decreases) > 5 {
		decreases = decreases[:5]
	}

	// Average change
	avgChange := decimal.Zero
	if changeCount > 0 {
		avgChange = changeSum.Div(decimal.NewFromInt(int64(changeCount)))
	}

	// Store ranking: for each item, which store has the lowest latest unit price
	storeWins := make(map[string]int)
	for _, name := range names {
		nameItems := byName[name]
		// Get the latest entry per store
		latestByStore := make(map[string]decimal.Decimal)
		for _, item := range nameItems {
			if _, exists := latestByStore[item.Store]; !exists {
				latestByStore[item.Store] = item.UnitPrice
			}
		}
		// Find cheapest store
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

	// Recent entries: group by date+store, count items
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

	// Collect distinct filter values
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

	// 90-day change
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

// Helper functions

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
```

- [ ] **Step 2: Register GET routes in server.go**

Add these routes in `internal/server/server.go` after the POST route from Task 2:

```go
router.GET("/api/price_tracking", func(c *gin.Context) {
	c.JSON(200, GetPriceTracking(db))
})

router.GET("/api/price_tracking/items", func(c *gin.Context) {
	c.JSON(200, GetPriceTrackingItems(db))
})

router.GET("/api/price_tracking/item/:name", func(c *gin.Context) {
	name := c.Param("name")
	c.JSON(200, GetPriceTrackingItem(db, name))
})
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/server/price_tracking.go internal/server/server.go
git commit -m "feat(price-tracking): add GET endpoints for dashboard and item detail"
```

---

### Task 4: Backend Integration Test

**Files:**
- Modify: `tests/regression.test.ts` (reference only — test manually)

- [ ] **Step 1: Build and start the server**

Run: `go build`
Expected: Binary built successfully

- [ ] **Step 2: Manually test the POST endpoint**

Start the server in background and test:

```bash
./paisa serve &
sleep 2

# POST a receipt
curl -X POST http://localhost:7500/api/price_tracking/receipt \
  -H "Content-Type: application/json" \
  -d '{
    "store": "Costco",
    "date": "2026-03-28",
    "items": [
      {"name": "milk", "brand": "Kirkland", "variant": "Whole, Organic", "unit": "gal", "quantity": 2, "price": 9.78},
      {"name": "eggs", "brand": "Happy Egg", "variant": "Free Range, Large", "unit": "doz", "quantity": 1, "price": 5.49}
    ]
  }'
```

Expected: `{"count":2,"success":true}`

- [ ] **Step 3: Test GET endpoints**

```bash
# Dashboard
curl http://localhost:7500/api/price_tracking

# Items list
curl http://localhost:7500/api/price_tracking/items

# Item detail
curl http://localhost:7500/api/price_tracking/item/milk
```

Expected: JSON responses with the posted data reflected.

- [ ] **Step 4: Stop the server and commit if any fixes were needed**

```bash
pkill -f 'paisa serve'
```

---

### Task 5: Frontend Types and Utility Module

**Files:**
- Create: `src/lib/price_tracking.ts`

- [ ] **Step 1: Create the price tracking frontend module**

Create `src/lib/price_tracking.ts`:

```typescript
import * as d3 from "d3";
import _ from "lodash";
import { formatCurrency, type Legend } from "./utils";
import { generateColorScheme } from "./colors";

export interface ReceiptItem {
  id: number;
  date: string;
  store: string;
  name: string;
  brand: string;
  variant: string;
  unit: string;
  quantity: number;
  price: number;
  unit_price: number;
}

export interface PriceMover {
  name: string;
  unit: string;
  change: number;
  latest_price: number;
}

export interface StoreRanking {
  store: string;
  count: number;
}

export interface RecentEntry {
  date: string;
  store: string;
  count: number;
}

export interface PriceTrackingDashboard {
  items_count: number;
  stores_count: number;
  avg_change: number;
  increases: PriceMover[];
  decreases: PriceMover[];
  store_ranking: StoreRanking[];
  recent: RecentEntry[];
}

export interface PriceTrackingItemDetail {
  name: string;
  entries: ReceiptItem[];
  stores: string[];
  brands: string[];
  variants: string[];
  latest: ReceiptItem | null;
  avg: number;
  change: number;
}

export function formatChange(change: number): string {
  const sign = change >= 0 ? "▲" : "▼";
  return `${sign} ${Math.abs(change).toFixed(1)}%`;
}

export function changeColor(change: number): string {
  if (change > 0) return "#e74c3c";
  if (change < 0) return "#27ae60";
  return "#888";
}

export function renderPriceTimeline(
  id: string,
  entries: ReceiptItem[],
  filterStore: string,
  filterBrand: string,
  filterVariant: string
): { destroy: () => void; legends: Legend[] } {
  const el = document.getElementById(id.substring(1));
  if (!el) return { destroy: () => {}, legends: [] };

  const svg = d3.select(id);
  svg.selectAll("*").remove();

  let filtered = entries;
  if (filterStore) filtered = filtered.filter((e) => e.store === filterStore);
  if (filterBrand) filtered = filtered.filter((e) => e.brand === filterBrand);
  if (filterVariant) filtered = filtered.filter((e) => e.variant === filterVariant);

  if (filtered.length === 0) {
    return { destroy: () => {}, legends: [] };
  }

  const margin = { top: 20, right: 30, bottom: 40, left: 60 };
  const width = el.parentElement.clientWidth - margin.left - margin.right;
  const height = +svg.attr("height") - margin.top - margin.bottom;
  const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);

  const parseDate = (d: string) => new Date(d);

  const byStore = _.groupBy(filtered, "store");
  const stores = Object.keys(byStore);
  const colors = generateColorScheme(stores);

  const allDates = filtered.map((e) => parseDate(e.date));
  const allPrices = filtered.map((e) => e.unit_price);

  const x = d3
    .scaleTime()
    .domain(d3.extent(allDates) as [Date, Date])
    .range([0, width]);

  const y = d3
    .scaleLinear()
    .domain([0, d3.max(allPrices) * 1.1])
    .range([height, 0]);

  g.append("g").attr("transform", `translate(0,${height})`).call(d3.axisBottom(x));

  g.append("g").call(d3.axisLeft(y).tickFormat((d) => formatCurrency(d as number)));

  const line = d3
    .line<ReceiptItem>()
    .x((d) => x(parseDate(d.date)))
    .y((d) => y(d.unit_price));

  const legends: Legend[] = [];

  stores.forEach((store, i) => {
    const storeItems = _.sortBy(byStore[store], "date");
    const color = colors(store);

    g.append("path")
      .datum(storeItems)
      .attr("fill", "none")
      .attr("stroke", color)
      .attr("stroke-width", 2)
      .attr("d", line);

    g.selectAll(`.dot-${i}`)
      .data(storeItems)
      .join("circle")
      .attr("cx", (d) => x(parseDate(d.date)))
      .attr("cy", (d) => y(d.unit_price))
      .attr("r", 4)
      .attr("fill", color);

    legends.push({ label: store, color, shape: "circle" });
  });

  return {
    destroy: () => svg.selectAll("*").remove(),
    legends,
  };
}
```

- [ ] **Step 2: Verify the frontend builds**

Run: `npm run build`
Expected: No errors (the module is created but not yet imported anywhere)

- [ ] **Step 3: Commit**

```bash
git add src/lib/price_tracking.ts
git commit -m "feat(price-tracking): add frontend types and D3 price timeline renderer"
```

---

### Task 6: Dashboard Page

**Files:**
- Create: `src/routes/(app)/more/price_tracking/+page.svelte`
- Modify: `src/lib/components/Navbar.svelte`

- [ ] **Step 1: Create the dashboard page**

Create `src/routes/(app)/more/price_tracking/+page.svelte`:

```svelte
<script lang="ts">
  import {
    ajax,
    formatCurrency,
    isMobile
  } from "$lib/utils";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import {
    formatChange,
    changeColor,
    type PriceTrackingDashboard,
    type PriceMover,
    type StoreRanking,
    type RecentEntry
  } from "$lib/price_tracking";
  import { onMount } from "svelte";
  import dayjs from "dayjs";

  let dashboard: PriceTrackingDashboard = null;
  let isEmpty = true;

  onMount(async () => {
    dashboard = await ajax("/api/price_tracking");
    isEmpty = dashboard.items_count === 0;
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap is-centered">
      <ZeroState item={!isEmpty}>
        <strong>No price data yet.</strong> Post receipt data to the
        <code>/api/price_tracking/receipt</code> endpoint to start tracking prices.
      </ZeroState>

      {#if dashboard && !isEmpty}
        <div class="column is-12">
          <nav class="level {isMobile() && 'grid-2'}">
            <LevelItem title="Items Tracked" value={String(dashboard.items_count)} />
            <LevelItem title="Stores" value={String(dashboard.stores_count)} />
            <LevelItem
              title="Avg Price Change (30d)"
              value={formatChange(dashboard.avg_change)}
              color={changeColor(dashboard.avg_change)}
            />
          </nav>
        </div>

        <div class="column is-6">
          <div class="box has-background-white">
            <p class="heading mb-3">Biggest Increases (30d)</p>
            {#each dashboard.increases as mover (mover.name)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <a class="secondary-link" href="/more/price_tracking/item/{mover.name}">
                  {mover.name} (1 {mover.unit})
                </a>
                <span style="color: {changeColor(mover.change)}">{formatChange(mover.change)}</span>
              </div>
            {/each}
            {#if dashboard.increases.length === 0}
              <p class="has-text-grey">Not enough data</p>
            {/if}
          </div>
        </div>

        <div class="column is-6">
          <div class="box has-background-white">
            <p class="heading mb-3">Biggest Decreases (30d)</p>
            {#each dashboard.decreases as mover (mover.name)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <a class="secondary-link" href="/more/price_tracking/item/{mover.name}">
                  {mover.name} (1 {mover.unit})
                </a>
                <span style="color: {changeColor(mover.change)}">{formatChange(mover.change)}</span>
              </div>
            {/each}
            {#if dashboard.decreases.length === 0}
              <p class="has-text-grey">Not enough data</p>
            {/if}
          </div>
        </div>

        <div class="column is-6">
          <div class="box has-background-white">
            <p class="heading mb-3">Cheapest Store by Item Count</p>
            {#each dashboard.store_ranking as ranking (ranking.store)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <span>{ranking.store}</span>
                <span class="has-text-weight-bold">{ranking.count} items cheapest</span>
              </div>
            {/each}
          </div>
        </div>

        <div class="column is-6">
          <div class="box has-background-white">
            <p class="heading mb-3">Recently Added</p>
            {#each dashboard.recent as entry (entry.date + entry.store)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <span>{dayjs(entry.date).format("MMM D")}</span>
                <span>{entry.store}</span>
                <span>{entry.count} items</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  </div>
</section>

<style lang="scss">
  .border-bottom {
    border-bottom: 1px solid #eee;
    &:last-child {
      border-bottom: none;
    }
  }
</style>
```

- [ ] **Step 2: Add navigation entry**

In `src/lib/components/Navbar.svelte`, find the "More" menu's children array. Add "Price Tracking" before the existing entries:

Find this block in the `links` array:
```typescript
{
    label: "More",
    href: "/more",
    children: [
      { label: "Configuration", href: "/config", help: "config" },
```

Add `{ label: "Price Tracking", href: "/price_tracking" },` as the first child:
```typescript
{
    label: "More",
    href: "/more",
    children: [
      { label: "Price Tracking", href: "/price_tracking" },
      { label: "Configuration", href: "/config", help: "config" },
```

- [ ] **Step 3: Verify the frontend builds**

Run: `npm run build`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add src/routes/(app)/more/price_tracking/+page.svelte src/lib/components/Navbar.svelte
git commit -m "feat(price-tracking): add dashboard page and navigation entry"
```

---

### Task 7: Item Detail Page

**Files:**
- Create: `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`

- [ ] **Step 1: Create the item detail page**

Create `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`:

```svelte
<script lang="ts">
  import { page } from "$app/stores";
  import { ajax, formatCurrency, isMobile } from "$lib/utils";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import {
    formatChange,
    changeColor,
    renderPriceTimeline,
    type PriceTrackingItemDetail,
    type ReceiptItem
  } from "$lib/price_tracking";
  import { onMount } from "svelte";
  import dayjs from "dayjs";

  let detail: PriceTrackingItemDetail = null;
  let filterStore = "";
  let filterBrand = "";
  let filterVariant = "";
  let chartDestroy: () => void = null;

  $: name = $page.params.name;

  $: filteredEntries = detail
    ? detail.entries.filter((e) => {
        if (filterStore && e.store !== filterStore) return false;
        if (filterBrand && e.brand !== filterBrand) return false;
        if (filterVariant && e.variant !== filterVariant) return false;
        return true;
      })
    : [];

  function renderChart() {
    if (chartDestroy) chartDestroy();
    if (!detail || detail.entries.length === 0) return;
    const result = renderPriceTimeline(
      "#d3-price-timeline",
      detail.entries,
      filterStore,
      filterBrand,
      filterVariant
    );
    chartDestroy = result.destroy;
  }

  $: if (detail && filterStore !== undefined) {
    // Re-render chart when filters change (tick to ensure DOM is ready)
    setTimeout(renderChart, 0);
  }

  onMount(async () => {
    detail = await ajax(`/api/price_tracking/item/${name}`);
    setTimeout(renderChart, 0);
  });
</script>

<section class="section">
  <div class="container is-fluid">
    {#if detail}
      <div class="columns is-flex-wrap-wrap">
        <div class="column is-12">
          <h2 class="title is-4 mb-2">{detail.name}</h2>
          {#if detail.latest}
            <nav class="level {isMobile() && 'grid-2'}">
              <LevelItem
                title="Latest"
                value="{formatCurrency(detail.latest.unit_price)}/{detail.latest.unit}"
              />
              <LevelItem
                title="Average (90d)"
                value="{formatCurrency(detail.avg)}/{detail.latest.unit}"
              />
              <LevelItem
                title="Change (90d)"
                value={formatChange(detail.change)}
                color={changeColor(detail.change)}
              />
            </nav>
          {/if}
        </div>

        <div class="column is-12">
          <div class="is-flex is-flex-wrap-wrap" style="gap: 8px;">
            <div class="select is-small">
              <select bind:value={filterStore}>
                <option value="">All Stores</option>
                {#each detail.stores as store}
                  <option value={store}>{store}</option>
                {/each}
              </select>
            </div>
            <div class="select is-small">
              <select bind:value={filterBrand}>
                <option value="">All Brands</option>
                {#each detail.brands as brand}
                  <option value={brand}>{brand}</option>
                {/each}
              </select>
            </div>
            <div class="select is-small">
              <select bind:value={filterVariant}>
                <option value="">All Variants</option>
                {#each detail.variants as variant}
                  <option value={variant}>{variant}</option>
                {/each}
              </select>
            </div>
          </div>
        </div>

        <div class="column is-12">
          <svg id="d3-price-timeline" height="300" width="100%"></svg>
        </div>

        <div class="column is-12">
          <div class="box has-background-white p-0" style="overflow-x: auto;">
            <table class="table is-fullwidth is-hoverable mb-0">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Store</th>
                  <th>Brand</th>
                  <th>Variant</th>
                  <th class="has-text-right">Qty</th>
                  <th class="has-text-right">Unit Price</th>
                  <th class="has-text-right">Change</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredEntries as entry, i (entry.id)}
                  {@const prev = filteredEntries[i + 1]}
                  {@const pctChange =
                    prev && prev.unit_price !== 0
                      ? ((entry.unit_price - prev.unit_price) / prev.unit_price) * 100
                      : null}
                  <tr>
                    <td>{dayjs(entry.date).format("MMM D, YYYY")}</td>
                    <td>{entry.store}</td>
                    <td>{entry.brand}</td>
                    <td>{entry.variant}</td>
                    <td class="has-text-right">
                      {entry.quantity}
                      {entry.unit}
                    </td>
                    <td class="has-text-right">{formatCurrency(entry.unit_price)}</td>
                    <td class="has-text-right">
                      {#if pctChange !== null}
                        <span style="color: {changeColor(pctChange)}"
                          >{formatChange(pctChange)}</span
                        >
                      {:else}
                        <span class="has-text-grey">—</span>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {/if}
  </div>
</section>
```

- [ ] **Step 2: Verify the frontend builds**

Run: `npm run build`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add src/routes/(app)/more/price_tracking/item/
git commit -m "feat(price-tracking): add item detail page with chart and history table"
```

---

### Task 8: End-to-End Verification

**Files:** None (manual testing)

- [ ] **Step 1: Build everything**

Run: `make jsbuild && go build`
Expected: Both frontend and backend build successfully

- [ ] **Step 2: Start the development server**

Run: `make develop`
Expected: Both Go and JS servers start

- [ ] **Step 3: Post test data via API**

In another terminal:

```bash
# Post receipt 1 — March 15
curl -X POST http://localhost:7500/api/price_tracking/receipt \
  -H "Content-Type: application/json" \
  -d '{
    "store": "Walmart",
    "date": "2026-03-15",
    "items": [
      {"name": "milk", "brand": "Great Value", "variant": "Whole", "unit": "gal", "quantity": 1, "price": 4.29},
      {"name": "eggs", "brand": "Egglands Best", "variant": "Large", "unit": "doz", "quantity": 1, "price": 4.99}
    ]
  }'

# Post receipt 2 — March 28
curl -X POST http://localhost:7500/api/price_tracking/receipt \
  -H "Content-Type: application/json" \
  -d '{
    "store": "Costco",
    "date": "2026-03-28",
    "items": [
      {"name": "milk", "brand": "Kirkland", "variant": "Whole, Organic", "unit": "gal", "quantity": 2, "price": 9.78},
      {"name": "eggs", "brand": "Happy Egg", "variant": "Free Range, Large", "unit": "doz", "quantity": 2, "price": 10.98}
    ]
  }'
```

Expected: Both return `{"count":2,"success":true}`

- [ ] **Step 4: Verify in browser**

Open `http://localhost:5173/more/price_tracking` and verify:
- Dashboard shows 2 items tracked, 2 stores
- Recent entries show both receipts
- Click on "milk" to navigate to item detail page
- Item detail shows price history chart and table with both entries
- Filters for store, brand, and variant work

- [ ] **Step 5: Run existing tests to check for regressions**

Run: `make lint && make test`
Expected: All existing tests pass, no regressions introduced

- [ ] **Step 6: Final commit if any fixes were needed**

```bash
git add -A
git commit -m "fix(price-tracking): address issues found during e2e testing"
```
