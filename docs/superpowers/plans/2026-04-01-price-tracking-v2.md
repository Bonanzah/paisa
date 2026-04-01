# Price Tracking v2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Improve the price tracking feature with a redesigned dashboard, browsable items list, polished D3 charts, store comparisons, inline edit/delete, and data normalization.

**Architecture:** Backend-first approach — add new Go endpoints and normalization logic, then update the frontend. The existing `receipt_items` table schema is unchanged. All new insights (store comparison, item summaries) are computed server-side or client-side from existing data.

**Tech Stack:** Go (Gin, GORM, shopspring/decimal), SvelteKit (Svelte 4), D3.js, Bulma + DaisyUI, TypeScript

---

## File Structure

**Backend (Go):**
- Modify: `internal/model/receipt_item/receipt_item.go` — add `FindByID`, `Update`, `Delete` model methods
- Modify: `internal/server/price_tracking.go` — add `normalize()` helper, `GetPriceTrackingItemsSummary()`, `UpdateReceiptItem()`, `DeleteReceiptItem()` handlers; apply normalization in `PostReceipt()`
- Modify: `internal/server/server.go` — register 3 new routes

**Frontend (Svelte/TypeScript):**
- Modify: `src/lib/price_tracking.ts` — add `ItemSummary` type, update `renderPriceTimeline()` with tooltips/smooth curves/legend toggling
- Modify: `src/routes/(app)/more/price_tracking/+page.svelte` — redesigned dashboard + items tab
- Modify: `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte` — store comparison section, inline edit/delete

---

### Task 1: Add Model Helper Methods (FindByID, Update, Delete)

**Files:**
- Modify: `internal/model/receipt_item/receipt_item.go`

- [ ] **Step 1: Add FindByID method**

Add to `internal/model/receipt_item/receipt_item.go` after the `DistinctStores` function:

```go
func FindByID(db *gorm.DB, id uint) (*ReceiptItem, error) {
	var item ReceiptItem
	result := db.First(&item, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &item, nil
}
```

- [ ] **Step 2: Add Update method**

Add after `FindByID`:

```go
func Update(db *gorm.DB, item *ReceiptItem) error {
	return db.Save(item).Error
}
```

- [ ] **Step 3: Add Delete method**

Add after `Update`:

```go
func Delete(db *gorm.DB, id uint) error {
	return db.Delete(&ReceiptItem{}, id).Error
}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && go build ./internal/model/receipt_item/`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add internal/model/receipt_item/receipt_item.go
git commit -m "feat(price-tracking): add FindByID, Update, Delete model methods"
```

---

### Task 2: Add Data Normalization and Apply to PostReceipt

**Files:**
- Modify: `internal/server/price_tracking.go`

- [ ] **Step 1: Add normalize helper functions**

Add these imports to `internal/server/price_tracking.go` — add `"strings"` and `"unicode"` to the import block. Then add the following functions after the existing `setToSlice` function at the bottom of the file:

```go
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
```

- [ ] **Step 2: Apply normalization in PostReceipt**

In `internal/server/price_tracking.go`, in the `PostReceipt` function, replace the item construction block (the `receipt_item.ReceiptItem{...}` struct literal inside the for loop, approximately lines 41-51) with:

```go
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
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && go build ./internal/server/`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/server/price_tracking.go
git commit -m "feat(price-tracking): add data normalization on receipt POST"
```

---

### Task 3: Add PUT and DELETE Endpoints for Receipt Items

**Files:**
- Modify: `internal/server/price_tracking.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Add UpdateReceiptItem handler**

Add to `internal/server/price_tracking.go` after the `GetPriceTrackingItem` function:

```go
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
```

- [ ] **Step 2: Add DeleteReceiptItem handler**

Add after `UpdateReceiptItem`:

```go
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
```

- [ ] **Step 3: Register routes in server.go**

In `internal/server/server.go`, add these routes after the existing `router.GET("/api/price_tracking/item/:name", ...)` block (after line 398). You will need to add `"strconv"` to the import block:

```go
	router.PUT("/api/price_tracking/receipt_item/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
			return
		}
		var request UpdateReceiptItemRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		result, status := UpdateReceiptItem(db, uint(id), request)
		c.JSON(status, result)
	})

	router.DELETE("/api/price_tracking/receipt_item/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
			return
		}
		result, status := DeleteReceiptItem(db, uint(id))
		c.JSON(status, result)
	})
```

- [ ] **Step 4: Verify it compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && go build ./...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add internal/server/price_tracking.go internal/server/server.go
git commit -m "feat(price-tracking): add PUT and DELETE endpoints for receipt items"
```

---

### Task 4: Add Items Summary Endpoint

**Files:**
- Modify: `internal/server/price_tracking.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Add ItemSummary type and GetPriceTrackingItemsSummary handler**

Add to `internal/server/price_tracking.go` after the `RecentEntry` type:

```go
type ItemSummary struct {
	Name          string          `json:"name"`
	Unit          string          `json:"unit"`
	LatestPrice   decimal.Decimal `json:"latest_price"`
	AvgPrice      decimal.Decimal `json:"avg_price"`
	Change        decimal.Decimal `json:"change"`
	StoresCount   int             `json:"stores_count"`
	LastPurchased time.Time       `json:"last_purchased"`
}
```

Add the handler function after `DeleteReceiptItem`:

```go
func GetPriceTrackingItemsSummary(db *gorm.DB) gin.H {
	items := receipt_item.All(db)
	names := receipt_item.DistinctNames(db)

	if len(items) == 0 {
		return gin.H{"items": []ItemSummary{}}
	}

	byName := make(map[string][]receipt_item.ReceiptItem)
	for _, item := range items {
		byName[item.Name] = append(byName[item.Name], item)
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)

	var summaries []ItemSummary
	for _, name := range names {
		nameItems := byName[name]
		if len(nameItems) == 0 {
			continue
		}

		latest := nameItems[0]

		// 90-day average
		sum := decimal.Zero
		count := 0
		for _, item := range nameItems {
			if item.Date.After(ninetyDaysAgo) {
				sum = sum.Add(item.UnitPrice)
				count++
			}
		}
		avgPrice := decimal.Zero
		if count > 0 {
			avgPrice = sum.Div(decimal.NewFromInt(int64(count)))
		}

		// 30-day change
		change := decimal.Zero
		if len(nameItems) >= 2 {
			var baseline *receipt_item.ReceiptItem
			for i := range nameItems {
				if nameItems[i].Date.Before(thirtyDaysAgo) {
					baseline = &nameItems[i]
					break
				}
			}
			if baseline != nil && !baseline.UnitPrice.IsZero() {
				change = latest.UnitPrice.Sub(baseline.UnitPrice).Div(baseline.UnitPrice).Mul(decimal.NewFromInt(100))
			}
		}

		// Distinct stores
		storeSet := make(map[string]bool)
		for _, item := range nameItems {
			storeSet[item.Store] = true
		}

		summaries = append(summaries, ItemSummary{
			Name:          name,
			Unit:          latest.Unit,
			LatestPrice:   latest.UnitPrice,
			AvgPrice:      avgPrice,
			Change:        change,
			StoresCount:   len(storeSet),
			LastPurchased: latest.Date,
		})
	}

	return gin.H{"items": summaries}
}
```

- [ ] **Step 2: Register route in server.go**

In `internal/server/server.go`, add this route before the existing `router.GET("/api/price_tracking/items", ...)` line (the new `/items/summary` route must come before `/items` or Gin may match `/items` first due to wildcard overlap — actually in Gin both are exact, so add it right after the existing `/items` route):

```go
	router.GET("/api/price_tracking/items/summary", func(c *gin.Context) {
		c.JSON(200, GetPriceTrackingItemsSummary(db))
	})
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/server/price_tracking.go internal/server/server.go
git commit -m "feat(price-tracking): add items summary endpoint"
```

---

### Task 5: Add Frontend Types and Ajax Overloads

**Files:**
- Modify: `src/lib/price_tracking.ts`
- Modify: `src/lib/utils.ts`

- [ ] **Step 1: Add ItemSummary type**

In `src/lib/price_tracking.ts`, add after the `PriceTrackingItemDetail` interface:

```typescript
export interface ItemSummary {
  name: string;
  unit: string;
  latest_price: number;
  avg_price: number;
  change: number;
  stores_count: number;
  last_purchased: string;
}
```

- [ ] **Step 2: Add ajax overloads in utils.ts**

In `src/lib/utils.ts`, find the existing price tracking ajax overloads (the block starting with `export function ajax(route: "/api/price_tracking"):`) and add these new overloads immediately after the existing `/api/price_tracking/item/:name` overload:

```typescript
export function ajax(route: "/api/price_tracking/items/summary"): Promise<{
  items: import("$lib/price_tracking").ItemSummary[];
}>;

export function ajax(
  route: "/api/price_tracking/receipt_item/:id",
  options?: RequestOptions,
  params?: Record<string, string>
): Promise<{ success: boolean; item?: any; error?: string }>;
```

- [ ] **Step 3: Verify frontend compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && npm run check`
Expected: No type errors

- [ ] **Step 4: Commit**

```bash
git add src/lib/price_tracking.ts src/lib/utils.ts
git commit -m "feat(price-tracking): add ItemSummary type and ajax overloads"
```

---

### Task 6: Improve D3 Chart (Tooltips, Smooth Curves, Legend Toggling)

**Files:**
- Modify: `src/lib/price_tracking.ts`

- [ ] **Step 1: Update renderPriceTimeline function**

In `src/lib/price_tracking.ts`, replace the entire `renderPriceTimeline` function (from `export function renderPriceTimeline(` through the closing `}` of the function, lines 69-152) with:

```typescript
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

  g.append("g")
    .attr("transform", `translate(0,${height})`)
    .call(d3.axisBottom(x).ticks(d3.timeMonth.every(1)).tickFormat(d3.timeFormat("%b %Y")));

  g.append("g").call(
    d3.axisLeft(y).tickFormat((d) => formatCurrency(d as number))
  );

  const line = d3
    .line<ReceiptItem>()
    .curve(d3.curveMonotoneX)
    .x((d) => x(parseDate(d.date)))
    .y((d) => y(d.unit_price));

  // Tooltip div
  let tooltip = d3.select("body").select(".pt-tooltip");
  if (tooltip.empty()) {
    tooltip = d3
      .select("body")
      .append("div")
      .attr("class", "pt-tooltip")
      .style("position", "absolute")
      .style("pointer-events", "none")
      .style("background", "rgba(0,0,0,0.8)")
      .style("color", "#fff")
      .style("padding", "6px 10px")
      .style("border-radius", "4px")
      .style("font-size", "12px")
      .style("display", "none")
      .style("z-index", "1000");
  }

  const hiddenStores = new Set<string>();
  const legends: Legend[] = [];

  stores.forEach((store, i) => {
    const storeItems = _.sortBy(byStore[store], "date");
    const color = colors(store);

    const pathEl = g
      .append("path")
      .datum(storeItems)
      .attr("fill", "none")
      .attr("stroke", color)
      .attr("stroke-width", 2)
      .attr("class", `line-${i}`)
      .attr("d", line);

    const dotsEl = g
      .selectAll(`.dot-${i}`)
      .data(storeItems)
      .join("circle")
      .attr("class", `dot-${i}`)
      .attr("cx", (d) => x(parseDate(d.date)))
      .attr("cy", (d) => y(d.unit_price))
      .attr("r", 4)
      .attr("fill", color)
      .on("mouseenter", (event: MouseEvent, d: ReceiptItem) => {
        tooltip
          .style("display", "block")
          .html(
            `<strong>${d.store}</strong><br/>` +
              `${formatCurrency(d.unit_price)}/${d.unit}<br/>` +
              `${new Date(d.date).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}` +
              (d.variant ? `<br/>${d.variant}` : "")
          );
      })
      .on("mousemove", (event: MouseEvent) => {
        tooltip
          .style("left", event.pageX + 12 + "px")
          .style("top", event.pageY - 10 + "px");
      })
      .on("mouseleave", () => {
        tooltip.style("display", "none");
      });

    legends.push({
      label: store,
      color,
      shape: "circle",
      toggle: () => {
        if (hiddenStores.has(store)) {
          hiddenStores.delete(store);
          pathEl.style("display", null);
          dotsEl.style("display", null);
        } else {
          hiddenStores.add(store);
          pathEl.style("display", "none");
          dotsEl.style("display", "none");
        }
      }
    });
  });

  return {
    destroy: () => {
      svg.selectAll("*").remove();
      tooltip.style("display", "none");
    },
    legends
  };
}
```

- [ ] **Step 2: Update Legend type if needed**

Check if the `Legend` type in `src/lib/utils.ts` has a `toggle` field. If not, find the `Legend` interface in `src/lib/utils.ts` and add an optional `toggle` field:

```typescript
toggle?: () => void;
```

- [ ] **Step 3: Verify frontend compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && npm run check`
Expected: No type errors

- [ ] **Step 4: Commit**

```bash
git add src/lib/price_tracking.ts src/lib/utils.ts
git commit -m "feat(price-tracking): add tooltips, smooth curves, legend toggling to D3 chart"
```

---

### Task 7: Redesign Dashboard Page with Tabs

**Files:**
- Modify: `src/routes/(app)/more/price_tracking/+page.svelte`

- [ ] **Step 1: Rewrite the dashboard page**

Replace the entire contents of `src/routes/(app)/more/price_tracking/+page.svelte` with:

```svelte
<script lang="ts">
  import { ajax, formatCurrency, isMobile } from "$lib/utils";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import {
    formatChange,
    changeColor,
    type PriceTrackingDashboard,
    type ItemSummary
  } from "$lib/price_tracking";
  import { onMount } from "svelte";
  import dayjs from "dayjs";

  let dashboard: PriceTrackingDashboard = null;
  let itemsSummary: ItemSummary[] = [];
  let isEmpty = true;
  let activeTab: "dashboard" | "items" = "dashboard";

  // Items tab state
  let searchQuery = "";
  let sortBy: "name" | "latest_price" | "change" = "name";
  let sortAsc = true;

  $: filteredItems = itemsSummary
    .filter((item) => item.name.toLowerCase().includes(searchQuery.toLowerCase()))
    .sort((a, b) => {
      let cmp = 0;
      if (sortBy === "name") cmp = a.name.localeCompare(b.name);
      else if (sortBy === "latest_price") cmp = a.latest_price - b.latest_price;
      else if (sortBy === "change") cmp = a.change - b.change;
      return sortAsc ? cmp : -cmp;
    });

  function toggleSort(col: "name" | "latest_price" | "change") {
    if (sortBy === col) {
      sortAsc = !sortAsc;
    } else {
      sortBy = col;
      sortAsc = col === "name";
    }
  }

  onMount(async () => {
    const [dash, summary] = await Promise.all([
      ajax("/api/price_tracking"),
      ajax("/api/price_tracking/items/summary")
    ]);
    dashboard = dash;
    itemsSummary = summary.items || [];
    isEmpty = dashboard.items_count === 0;
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <ZeroState item={!isEmpty}>
      <strong>No price data yet.</strong> Post receipt data to the
      <code>/api/price_tracking/receipt</code> endpoint to start tracking prices.
    </ZeroState>

    {#if dashboard && !isEmpty}
      <div class="du-tabs du-tabs-bordered mb-5">
        <button
          class="du-tab"
          class:du-tab-active={activeTab === "dashboard"}
          on:click={() => (activeTab = "dashboard")}>Dashboard</button
        >
        <button
          class="du-tab"
          class:du-tab-active={activeTab === "items"}
          on:click={() => (activeTab = "items")}>Items</button
        >
      </div>

      {#if activeTab === "dashboard"}
        <nav class="level {isMobile() && 'grid-2'} mb-5">
          <LevelItem title="Items Tracked" value={String(dashboard.items_count)} />
          <LevelItem title="Stores" value={String(dashboard.stores_count)} />
          <LevelItem
            title="Avg Price Change (30d)"
            value={formatChange(dashboard.avg_change)}
            color={changeColor(dashboard.avg_change)}
          />
        </nav>

        <div class="columns">
          <div class="column is-6">
            <p class="is-size-6 has-text-weight-semibold has-text-grey mb-3">Price Movers (30d)</p>

            <p class="is-size-7 has-text-weight-bold has-text-grey-dark mb-2">INCREASES</p>
            {#each dashboard.increases ?? [] as mover (mover.name)}
              <div class="is-flex is-justify-content-space-between is-align-items-center py-2 border-bottom">
                <a class="secondary-link" href="/more/price_tracking/item/{mover.name}">
                  {mover.name}
                </a>
                <span class="is-flex" style="gap: 12px;">
                  <span class="has-text-grey is-size-7">{formatCurrency(mover.latest_price)}/{mover.unit}</span>
                  <span class="has-text-weight-bold is-size-7" style="color: {changeColor(mover.change)}"
                    >{formatChange(mover.change)}</span
                  >
                </span>
              </div>
            {/each}
            {#if (dashboard.increases ?? []).length === 0}
              <p class="has-text-grey is-size-7">Not enough data</p>
            {/if}

            <p class="is-size-7 has-text-weight-bold has-text-grey-dark mb-2 mt-4">DECREASES</p>
            {#each dashboard.decreases ?? [] as mover (mover.name)}
              <div class="is-flex is-justify-content-space-between is-align-items-center py-2 border-bottom">
                <a class="secondary-link" href="/more/price_tracking/item/{mover.name}">
                  {mover.name}
                </a>
                <span class="is-flex" style="gap: 12px;">
                  <span class="has-text-grey is-size-7">{formatCurrency(mover.latest_price)}/{mover.unit}</span>
                  <span class="has-text-weight-bold is-size-7" style="color: {changeColor(mover.change)}"
                    >{formatChange(mover.change)}</span
                  >
                </span>
              </div>
            {/each}
            {#if (dashboard.decreases ?? []).length === 0}
              <p class="has-text-grey is-size-7">Not enough data</p>
            {/if}
          </div>

          <div class="column is-6">
            <p class="is-size-6 has-text-weight-semibold has-text-grey mb-3">Store Rankings</p>
            {#each dashboard.store_ranking ?? [] as ranking, i (ranking.store)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <span>
                  <span class="has-text-grey is-size-7 mr-2">#{i + 1}</span>
                  {ranking.store}
                </span>
                <span class="has-text-weight-bold is-size-7">{ranking.count} items cheapest</span>
              </div>
            {/each}

            <p class="is-size-6 has-text-weight-semibold has-text-grey mb-3 mt-5">Recently Added</p>
            {#each dashboard.recent ?? [] as entry (entry.date + entry.store)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <span class="has-text-grey is-size-7">{dayjs(entry.date).format("MMM D")}</span>
                <span>{entry.store}</span>
                <span class="is-size-7">{entry.count} items</span>
              </div>
            {/each}
          </div>
        </div>
      {:else if activeTab === "items"}
        <div class="mb-4">
          <div class="field">
            <div class="control has-icons-left">
              <input
                class="input"
                type="text"
                placeholder="Search items..."
                bind:value={searchQuery}
              />
              <span class="icon is-left">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              </span>
            </div>
          </div>
        </div>

        <div class="box has-background-white p-0" style="overflow-x: auto;">
          <table class="table is-fullwidth is-hoverable mb-0">
            <thead>
              <tr>
                <th class="is-clickable" on:click={() => toggleSort("name")}>
                  Name {sortBy === "name" ? (sortAsc ? "↑" : "↓") : ""}
                </th>
                <th class="has-text-right is-clickable" on:click={() => toggleSort("latest_price")}>
                  Latest Price {sortBy === "latest_price" ? (sortAsc ? "↑" : "↓") : ""}
                </th>
                <th class="has-text-right">90d Avg</th>
                <th class="has-text-right is-clickable" on:click={() => toggleSort("change")}>
                  30d Change {sortBy === "change" ? (sortAsc ? "↑" : "↓") : ""}
                </th>
                <th class="has-text-right">Stores</th>
                <th class="has-text-right">Last Purchased</th>
              </tr>
            </thead>
            <tbody>
              {#each filteredItems as item (item.name)}
                <tr>
                  <td>
                    <a class="secondary-link" href="/more/price_tracking/item/{item.name}">
                      {item.name}
                    </a>
                  </td>
                  <td class="has-text-right">
                    {formatCurrency(item.latest_price)}/{item.unit}
                  </td>
                  <td class="has-text-right">
                    {formatCurrency(item.avg_price)}/{item.unit}
                  </td>
                  <td class="has-text-right">
                    <span style="color: {changeColor(item.change)}">{formatChange(item.change)}</span>
                  </td>
                  <td class="has-text-right">{item.stores_count}</td>
                  <td class="has-text-right">{dayjs(item.last_purchased).format("MMM D, YYYY")}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        {#if filteredItems.length === 0 && searchQuery}
          <p class="has-text-grey has-text-centered mt-4">No items matching "{searchQuery}"</p>
        {/if}
      {/if}
    {/if}
  </div>
</section>

<style lang="scss">
  .border-bottom {
    border-bottom: 1px solid rgba(128, 128, 128, 0.2);
    &:last-child {
      border-bottom: none;
    }
  }
</style>
```

- [ ] **Step 2: Verify frontend compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && npm run check`
Expected: No errors

- [ ] **Step 3: Visually verify**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && make develop`
Open the price tracking page. Verify:
- Tabs switch between Dashboard and Items views
- Dashboard shows summary row + two-column layout with movers and rankings
- Items tab shows searchable, sortable table
- Clicking an item name navigates to the detail page

- [ ] **Step 4: Commit**

```bash
git add src/routes/\(app\)/more/price_tracking/+page.svelte
git commit -m "feat(price-tracking): redesign dashboard with tabs and items list"
```

---

### Task 8: Add Store Comparison Section to Item Detail Page

**Files:**
- Modify: `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`

- [ ] **Step 1: Add store comparison computed data and section**

In `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`, add these computed values in the `<script>` block after the `filteredEntries` reactive declaration:

```typescript
  interface StoreComparison {
    store: string;
    latestPrice: number;
    vsAvg: number;
    lastDate: string;
  }

  $: storeComparisons = (() => {
    if (!detail || detail.entries.length === 0) return [] as StoreComparison[];
    const latestByStore = new Map<string, { price: number; date: string }>();
    for (const e of detail.entries) {
      if (!latestByStore.has(e.store)) {
        latestByStore.set(e.store, { price: e.unit_price, date: e.date });
      }
    }
    const prices = Array.from(latestByStore.values()).map((v) => v.price);
    const avg = prices.reduce((a, b) => a + b, 0) / prices.length;
    const result: StoreComparison[] = [];
    for (const [store, data] of latestByStore) {
      result.push({
        store,
        latestPrice: data.price,
        vsAvg: avg !== 0 ? ((data.price - avg) / avg) * 100 : 0,
        lastDate: data.date
      });
    }
    result.sort((a, b) => a.latestPrice - b.latestPrice);
    return result;
  })();

  $: bestStore = storeComparisons.length > 0 ? storeComparisons[0] : null;
  $: avgAcrossStores =
    storeComparisons.length > 0
      ? storeComparisons.reduce((a, b) => a + b.latestPrice, 0) / storeComparisons.length
      : 0;
  $: bestSavings =
    bestStore && avgAcrossStores > 0
      ? ((avgAcrossStores - bestStore.latestPrice) / avgAcrossStores) * 100
      : 0;
```

Then, in the template, add this block after the filter bar `<div>` (after the closing `</div>` of the `is-flex is-flex-wrap-wrap` div at ~line 106) and before the chart `<svg>`:

```svelte
        {#if storeComparisons.length > 1}
          <div class="column is-12">
            {#if bestStore}
              <div class="notification is-light is-size-7 py-2 px-4 mb-3">
                Best price: <strong>{formatCurrency(bestStore.latestPrice)}/{detail.latest.unit}</strong>
                at <strong>{bestStore.store}</strong>
                {#if bestSavings > 0.5}
                  (saves {bestSavings.toFixed(0)}% vs avg)
                {/if}
              </div>
            {/if}

            <table class="table is-fullwidth is-size-7 mb-4">
              <thead>
                <tr>
                  <th>Store</th>
                  <th class="has-text-right">Latest Price</th>
                  <th class="has-text-right">vs Avg</th>
                  <th class="has-text-right">Last Seen</th>
                </tr>
              </thead>
              <tbody>
                {#each storeComparisons as sc (sc.store)}
                  <tr>
                    <td>{sc.store}</td>
                    <td class="has-text-right">{formatCurrency(sc.latestPrice)}/{detail.latest.unit}</td>
                    <td class="has-text-right">
                      <span style="color: {changeColor(sc.vsAvg)}">{formatChange(sc.vsAvg)}</span>
                    </td>
                    <td class="has-text-right">{dayjs(sc.lastDate).format("MMM D, YYYY")}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
```

- [ ] **Step 2: Verify frontend compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && npm run check`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add src/routes/\(app\)/more/price_tracking/item/\[name\]/+page.svelte
git commit -m "feat(price-tracking): add store comparison section to item detail"
```

---

### Task 9: Add Inline Edit and Delete to Item Detail Page

**Files:**
- Modify: `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`

- [ ] **Step 1: Add edit/delete state and handlers**

In the `<script>` block of `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`, add these variables after the existing `let` declarations:

```typescript
  let editingId: number | null = null;
  let editForm = { store: "", brand: "", variant: "", quantity: "", price: "" };
  let deleteConfirmId: number | null = null;

  function startEdit(entry: ReceiptItem) {
    editingId = entry.id;
    editForm = {
      store: entry.store,
      brand: entry.brand,
      variant: entry.variant,
      quantity: String(entry.quantity),
      price: String(entry.price)
    };
  }

  function cancelEdit() {
    editingId = null;
  }

  async function saveEdit(id: number) {
    await ajax("/api/price_tracking/receipt_item/:id", {
      method: "PUT",
      body: JSON.stringify({
        store: editForm.store,
        brand: editForm.brand,
        variant: editForm.variant,
        quantity: parseFloat(editForm.quantity),
        price: parseFloat(editForm.price)
      })
    }, { id: String(id) });
    editingId = null;
    detail = await ajax("/api/price_tracking/item/:name", null, { name });
    setTimeout(renderChart, 0);
  }

  async function deleteEntry(id: number) {
    await ajax("/api/price_tracking/receipt_item/:id", {
      method: "DELETE"
    }, { id: String(id) });
    deleteConfirmId = null;
    detail = await ajax("/api/price_tracking/item/:name", null, { name });
    setTimeout(renderChart, 0);
  }
```

Also add `type ReceiptItem` to the import from `$lib/price_tracking` at the top:

```typescript
  import {
    formatChange,
    changeColor,
    renderPriceTimeline,
    type PriceTrackingItemDetail,
    type ReceiptItem
  } from "$lib/price_tracking";
```

- [ ] **Step 2: Update the history table with edit/delete UI**

Replace the `<tbody>` block in the history table (the entire `<tbody>...</tbody>` section) with:

```svelte
              <tbody>
                {#each filteredEntries as entry, i (entry.id)}
                  {@const prev = filteredEntries[i + 1]}
                  {@const pctChange =
                    prev && prev.unit_price !== 0
                      ? ((entry.unit_price - prev.unit_price) / prev.unit_price) * 100
                      : null}
                  {#if editingId === entry.id}
                    <tr>
                      <td>{dayjs(entry.date).format("MMM D, YYYY")}</td>
                      <td><input class="input is-small" bind:value={editForm.store} /></td>
                      <td><input class="input is-small" bind:value={editForm.brand} /></td>
                      <td><input class="input is-small" bind:value={editForm.variant} /></td>
                      <td class="has-text-right">
                        <input
                          class="input is-small"
                          type="number"
                          step="any"
                          bind:value={editForm.quantity}
                          style="width: 70px; display: inline-block;"
                        />
                        {entry.unit}
                      </td>
                      <td class="has-text-right">
                        <input
                          class="input is-small"
                          type="number"
                          step="any"
                          bind:value={editForm.price}
                          style="width: 80px; display: inline-block;"
                        />
                      </td>
                      <td class="has-text-right">
                        <button class="button is-small is-success mr-1" on:click={() => saveEdit(entry.id)}>Save</button>
                        <button class="button is-small" on:click={cancelEdit}>Cancel</button>
                      </td>
                    </tr>
                  {:else}
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
                        {#if deleteConfirmId === entry.id}
                          <span class="is-size-7 mr-2">Delete?</span>
                          <button class="button is-small is-danger mr-1" on:click={() => deleteEntry(entry.id)}>Yes</button>
                          <button class="button is-small" on:click={() => (deleteConfirmId = null)}>No</button>
                        {:else}
                          {#if pctChange !== null}
                            <span style="color: {changeColor(pctChange)}" class="mr-2"
                              >{formatChange(pctChange)}</span
                            >
                          {:else}
                            <span class="has-text-grey mr-2">—</span>
                          {/if}
                          <button class="button is-small is-ghost px-1" title="Edit" on:click={() => startEdit(entry)}>
                            ✎
                          </button>
                          <button class="button is-small is-ghost px-1 has-text-danger" title="Delete" on:click={() => (deleteConfirmId = entry.id)}>
                            ✕
                          </button>
                        {/if}
                      </td>
                    </tr>
                  {/if}
                {/each}
              </tbody>
```

- [ ] **Step 3: Update table header to add Actions column**

Replace the existing `<thead>` in the history table with:

```svelte
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Store</th>
                  <th>Brand</th>
                  <th>Variant</th>
                  <th class="has-text-right">Qty</th>
                  <th class="has-text-right">Unit Price</th>
                  <th class="has-text-right">Change / Actions</th>
                </tr>
              </thead>
```

- [ ] **Step 4: Verify frontend compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && npm run check`
Expected: No errors

- [ ] **Step 5: Visually verify**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && make develop`
Navigate to an item detail page. Verify:
- Edit icon converts row to inline form with Save/Cancel
- Delete icon shows confirmation with Yes/No
- Both actions refresh the data and chart

- [ ] **Step 6: Commit**

```bash
git add src/routes/\(app\)/more/price_tracking/item/\[name\]/+page.svelte
git commit -m "feat(price-tracking): add inline edit and delete to item detail"
```

---

### Task 10: Update Chart Legend Rendering on Item Detail Page

**Files:**
- Modify: `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`

- [ ] **Step 1: Add legend state and rendering**

In the `<script>` block, add a variable after the existing `let chartDestroy`:

```typescript
  import type { Legend } from "$lib/utils";
  let chartLegends: Legend[] = [];
```

Update the `renderChart` function to capture legends:

```typescript
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
    chartLegends = result.legends;
  }
```

- [ ] **Step 2: Add clickable legend below the chart SVG**

In the template, after the `<svg id="d3-price-timeline" ...></svg>` line, add:

```svelte
            {#if chartLegends.length > 1}
              <div class="is-flex is-flex-wrap-wrap mt-2" style="gap: 12px;">
                {#each chartLegends as legend}
                  <button
                    class="button is-small is-ghost px-2"
                    style="text-decoration: none;"
                    on:click={() => { if (legend.toggle) legend.toggle(); }}
                  >
                    <span
                      style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: {legend.color}; margin-right: 6px;"
                    ></span>
                    {legend.label}
                  </button>
                {/each}
              </div>
            {/if}
```

- [ ] **Step 3: Verify frontend compiles**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && npm run check`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add src/routes/\(app\)/more/price_tracking/item/\[name\]/+page.svelte
git commit -m "feat(price-tracking): add clickable chart legend to item detail"
```

---

### Task 11: Final Verification

- [ ] **Step 1: Run full lint**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && make lint`
Expected: No errors. If prettier formatting issues, run `npx prettier --write` on affected files and commit.

- [ ] **Step 2: Run Go build**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && go build ./...`
Expected: Clean build

- [ ] **Step 3: Run JS build**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && make jsbuild`
Expected: Clean build

- [ ] **Step 4: Run tests**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && make test`
Expected: All tests pass. Note: regression tests may need golden file updates if price tracking API responses changed shape — if so, run `make regen` and verify diffs make sense.

- [ ] **Step 5: Manual smoke test**

Run: `cd /Volumes/Develop/Repos/Aitchdien/paisa && make develop`
Then seed data: `bash scripts/seed_price_tracking.sh`

Verify:
1. Dashboard tab shows summary stats, price movers (increases/decreases in one card), store rankings, recent entries
2. Items tab shows searchable/sortable table with all items and stats
3. Clicking an item navigates to detail page
4. Detail page shows store comparison banner and table (when >1 store)
5. D3 chart has smooth curves, tooltips on hover, clickable legend
6. Edit button on history table opens inline form, save persists changes
7. Delete button shows confirmation, deleting removes the row
8. Data normalization: POST a receipt with "MILK" — verify it's stored as "milk"

- [ ] **Step 6: Fix any issues found and commit**

```bash
git add -A
git commit -m "fix(price-tracking): address issues from final verification"
```
