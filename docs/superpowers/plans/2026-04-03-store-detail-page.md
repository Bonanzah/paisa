# Store Detail Page & Back Button Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a store detail page showing average price trends and items purchased at a specific store, add contextual back buttons to detail pages, and make store names clickable throughout the app.

**Architecture:** New backend endpoint `GET /api/price_tracking/store/:name` returns store-level data. New Svelte page renders KPI metrics, a D3 line chart of average unit price over time, and a sortable/searchable items table with cross-store price comparison. The "vs Other Stores" comparison uses the existing `/api/price_tracking/items/summary` endpoint fetched client-side.

**Tech Stack:** Go/Gin backend, SvelteKit frontend (Svelte 4), D3.js charts, Bulma CSS, DaisyUI

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/model/receipt_item/receipt_item.go` | Modify | Add `ByStore()` query |
| `internal/server/price_tracking.go` | Modify | Add `GetPriceTrackingStore()` function |
| `internal/server/server.go` | Modify | Register new route |
| `src/lib/utils.ts` | Modify | Add ajax overload for store endpoint |
| `src/lib/price_tracking.ts` | Modify | Add `PriceTrackingStoreDetail` type + `renderStoreAvgPriceTimeline()` chart function |
| `src/routes/(app)/more/price_tracking/store/[name]/+page.svelte` | Create | Store detail page |
| `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte` | Modify | Add back button |
| `src/routes/(app)/more/price_tracking/+page.svelte` | Modify | Make store names clickable links |

---

### Task 1: Add `ByStore` model query

**Files:**
- Modify: `internal/model/receipt_item/receipt_item.go:45-52` (add after `ByName`)

- [ ] **Step 1: Add `ByStore` function**

Add this function after the existing `ByName` function (line 52) in `internal/model/receipt_item/receipt_item.go`:

```go
func ByStore(db *gorm.DB, store string) []ReceiptItem {
	var items []ReceiptItem
	result := db.Where("store = ?", store).Order("date DESC").Find(&items)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return items
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/model/receipt_item/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/model/receipt_item/receipt_item.go
git commit -m "feat(price-tracking): add ByStore query to receipt_item model"
```

---

### Task 2: Add `GetPriceTrackingStore` backend handler

**Files:**
- Modify: `internal/server/price_tracking.go` (add function after `GetPriceTrackingItem`, around line 287)
- Modify: `internal/server/server.go:403` (register route after item endpoint)

- [ ] **Step 1: Add `GetPriceTrackingStore` function**

Add this function after `GetPriceTrackingItem` in `internal/server/price_tracking.go`:

```go
func GetPriceTrackingStore(db *gorm.DB, storeName string) gin.H {
	items := receipt_item.ByStore(db, storeName)

	if len(items) == 0 {
		return gin.H{
			"store":          storeName,
			"entries":        []receipt_item.ReceiptItem{},
			"items":          []string{},
			"item_count":     0,
			"avg_unit_price": decimal.Zero,
			"change":         decimal.Zero,
			"last_purchased": nil,
		}
	}

	nameSet := make(map[string]bool)
	sum := decimal.Zero
	for _, item := range items {
		nameSet[item.Name] = true
		sum = sum.Add(item.UnitPrice)
	}

	avg := sum.Div(decimal.NewFromInt(int64(len(items))))

	// 90-day change: compare average unit price of latest entries vs entries from 90+ days ago
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	recentSum := decimal.Zero
	recentCount := 0
	oldSum := decimal.Zero
	oldCount := 0
	for _, item := range items {
		if item.Date.After(ninetyDaysAgo) {
			recentSum = recentSum.Add(item.UnitPrice)
			recentCount++
		} else {
			oldSum = oldSum.Add(item.UnitPrice)
			oldCount++
		}
	}

	change := decimal.Zero
	if recentCount > 0 && oldCount > 0 {
		recentAvg := recentSum.Div(decimal.NewFromInt(int64(recentCount)))
		oldAvg := oldSum.Div(decimal.NewFromInt(int64(oldCount)))
		if !oldAvg.IsZero() {
			change = recentAvg.Sub(oldAvg).Div(oldAvg).Mul(decimal.NewFromInt(100))
		}
	}

	return gin.H{
		"store":          storeName,
		"entries":        items,
		"items":          setToSlice(nameSet),
		"item_count":     len(nameSet),
		"avg_unit_price": avg,
		"change":         change,
		"last_purchased": items[0].Date,
	}
}
```

- [ ] **Step 2: Register the route in `server.go`**

In `internal/server/server.go`, add after the `GET /api/price_tracking/item/:name` route handler (after line 403):

```go
	router.GET("/api/price_tracking/store/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(200, GetPriceTrackingStore(db, name))
	})
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/server/price_tracking.go internal/server/server.go
git commit -m "feat(price-tracking): add GET /api/price_tracking/store/:name endpoint"
```

---

### Task 3: Add frontend types, ajax overload, and chart function

**Files:**
- Modify: `src/lib/price_tracking.ts` (add type + chart function)
- Modify: `src/lib/utils.ts` (add ajax overload)

- [ ] **Step 1: Add `PriceTrackingStoreDetail` type**

Add this interface after the existing `ItemSummary` interface (after line 66) in `src/lib/price_tracking.ts`:

```typescript
export interface PriceTrackingStoreDetail {
  store: string;
  entries: ReceiptItem[];
  items: string[];
  item_count: number;
  avg_unit_price: number;
  change: number;
  last_purchased: string;
}
```

- [ ] **Step 2: Add `renderStoreAvgPriceTimeline` chart function**

Add this function at the end of `src/lib/price_tracking.ts` (after the closing `}` of `renderPriceTimeline`):

```typescript
export function renderStoreAvgPriceTimeline(
  id: string,
  entries: ReceiptItem[]
): { destroy: () => void } {
  const el = document.getElementById(id.substring(1));
  if (!el) return { destroy: () => {} };

  const svg = d3.select(id);
  svg.selectAll("*").remove();

  if (entries.length === 0) {
    return { destroy: () => {} };
  }

  // Group entries by date and compute average unit price per date
  const byDate = _.groupBy(entries, (e) => e.date.substring(0, 10));
  const dateAvgs = Object.entries(byDate)
    .map(([date, items]) => ({
      date: new Date(date),
      avg: items.reduce((sum, e) => sum + e.unit_price, 0) / items.length
    }))
    .sort((a, b) => a.date.getTime() - b.date.getTime());

  const margin = { top: 20, right: 30, bottom: 40, left: 60 };
  const width = el.parentElement.clientWidth - margin.left - margin.right;
  const height = +svg.attr("height") - margin.top - margin.bottom;
  const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);

  const x = d3
    .scaleTime()
    .domain(d3.extent(dateAvgs, (d) => d.date) as [Date, Date])
    .range([0, width]);

  const y = d3
    .scaleLinear()
    .domain([0, d3.max(dateAvgs, (d) => d.avg) * 1.1])
    .range([height, 0]);

  g.append("g")
    .attr("transform", `translate(0,${height})`)
    .call(d3.axisBottom(x).ticks(d3.timeMonth.every(1)).tickFormat(d3.timeFormat("%b %Y")));

  g.append("g").call(d3.axisLeft(y).tickFormat((d) => formatCurrency(d as number)));

  const line = d3
    .line<{ date: Date; avg: number }>()
    .curve(d3.curveMonotoneX)
    .x((d) => x(d.date))
    .y((d) => y(d.avg));

  g.append("path")
    .datum(dateAvgs)
    .attr("fill", "none")
    .attr("stroke", "#3273dc")
    .attr("stroke-width", 2)
    .attr("d", line);

  // Tooltip
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

  g.selectAll(".dot")
    .data(dateAvgs)
    .join("circle")
    .attr("class", "dot")
    .attr("cx", (d) => x(d.date))
    .attr("cy", (d) => y(d.avg))
    .attr("r", 4)
    .attr("fill", "#3273dc")
    .on("mouseenter", (event: MouseEvent, d: { date: Date; avg: number }) => {
      tooltip.style("display", "block").html(
        `<strong>Avg Unit Price</strong><br/>` +
          `${formatCurrency(d.avg)}<br/>` +
          `${d.date.toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            year: "numeric"
          })}`
      );
    })
    .on("mousemove", (event: MouseEvent) => {
      tooltip.style("left", event.pageX + 12 + "px").style("top", event.pageY - 10 + "px");
    })
    .on("mouseleave", () => {
      tooltip.style("display", "none");
    });

  return {
    destroy: () => {
      svg.selectAll("*").remove();
      tooltip.style("display", "none");
    }
  };
}
```

- [ ] **Step 3: Add ajax overload for store endpoint**

In `src/lib/utils.ts`, add this overload before the catch-all `export async function ajax(route: string, ...)` implementation (before line 808):

```typescript
export function ajax(
  route: "/api/price_tracking/store/:name",
  options?: RequestOptions,
  params?: Record<string, string>
): Promise<import("$lib/price_tracking").PriceTrackingStoreDetail>;
```

- [ ] **Step 4: Verify frontend compiles**

Run: `npx svelte-check --threshold warning 2>&1 | head -20`
Expected: No new errors

- [ ] **Step 5: Commit**

```bash
git add src/lib/price_tracking.ts src/lib/utils.ts
git commit -m "feat(price-tracking): add store detail types, chart function, and ajax overload"
```

---

### Task 4: Create the store detail page

**Files:**
- Create: `src/routes/(app)/more/price_tracking/store/[name]/+page.svelte`

- [ ] **Step 1: Create the store detail page**

Create `src/routes/(app)/more/price_tracking/store/[name]/+page.svelte`:

```svelte
<script lang="ts">
  import { page } from "$app/stores";
  import { ajax, formatCurrency, isMobile } from "$lib/utils";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import {
    formatChange,
    changeColor,
    renderStoreAvgPriceTimeline,
    type PriceTrackingStoreDetail,
    type ItemSummary
  } from "$lib/price_tracking";
  import { onMount } from "svelte";
  import dayjs from "dayjs";

  let detail: PriceTrackingStoreDetail = null;
  let itemsSummary: ItemSummary[] = [];
  let chartDestroy: () => void = null;

  let searchQuery = "";
  let sortBy: "name" | "latest_price" | "change" = "name";
  let sortAsc = true;

  $: name = $page.params.name;

  interface StoreItemRow {
    name: string;
    unit: string;
    latestPrice: number;
    vsOtherStores: number | null;
    change: number;
    lastPurchased: string;
  }

  $: storeItems = (() => {
    if (!detail || detail.entries.length === 0) return [] as StoreItemRow[];

    const summaryByName = new Map(itemsSummary.map((s) => [s.name, s]));

    const byName = new Map<string, typeof detail.entries>();
    for (const e of detail.entries) {
      if (!byName.has(e.name)) byName.set(e.name, []);
      byName.get(e.name).push(e);
    }

    const rows: StoreItemRow[] = [];
    for (const [itemName, entries] of byName) {
      const latest = entries[0];
      const summary = summaryByName.get(itemName);

      // vs other stores: compare this store's latest price to global avg
      let vsOtherStores: number | null = null;
      if (summary && summary.stores_count > 1 && summary.avg_price > 0) {
        vsOtherStores = ((latest.unit_price - summary.avg_price) / summary.avg_price) * 100;
      }

      // 30d change at this store
      let change = 0;
      if (entries.length >= 2) {
        const thirtyDaysAgo = dayjs().subtract(30, "day");
        let baseline = null;
        for (let i = 0; i < entries.length; i++) {
          if (dayjs(entries[i].date).isBefore(thirtyDaysAgo)) {
            baseline = entries[i];
            break;
          }
        }
        if (baseline && baseline.unit_price !== 0) {
          change = ((latest.unit_price - baseline.unit_price) / baseline.unit_price) * 100;
        }
      }

      rows.push({
        name: itemName,
        unit: latest.unit,
        latestPrice: latest.unit_price,
        vsOtherStores,
        change,
        lastPurchased: latest.date
      });
    }
    return rows;
  })();

  $: filteredItems = storeItems
    .filter((item) => item.name.toLowerCase().includes(searchQuery.toLowerCase()))
    .sort((a, b) => {
      let cmp = 0;
      if (sortBy === "name") cmp = a.name.localeCompare(b.name);
      else if (sortBy === "latest_price") cmp = a.latestPrice - b.latestPrice;
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

  function renderChart() {
    if (chartDestroy) chartDestroy();
    if (!detail || detail.entries.length === 0) return;
    const result = renderStoreAvgPriceTimeline("#d3-store-price-timeline", detail.entries);
    chartDestroy = result.destroy;
  }

  onMount(async () => {
    const [storeDetail, summary] = await Promise.all([
      ajax("/api/price_tracking/store/:name", null, { name }),
      ajax("/api/price_tracking/items/summary")
    ]);
    detail = storeDetail;
    itemsSummary = summary.items || [];
    setTimeout(renderChart, 0);
  });
</script>

<section class="section">
  <div class="container is-fluid">
    {#if detail}
      <div class="columns is-flex-wrap-wrap">
        <div class="column is-12">
          <button
            class="button is-small is-ghost px-0 mb-3"
            style="text-decoration: none;"
            on:click={() => history.back()}
          >
            ← Back
          </button>
          <h2 class="title is-4 mb-2">{detail.store}</h2>
          <nav class="level {isMobile() && 'grid-2'}">
            <LevelItem title="Items Tracked" value={String(detail.item_count)} />
            <LevelItem
              title="Avg Unit Price"
              value={formatCurrency(detail.avg_unit_price)}
            />
            <LevelItem
              title="Price Trend (90d)"
              value={formatChange(detail.change)}
              color={changeColor(detail.change)}
            />
            <LevelItem
              title="Last Purchase"
              value={detail.last_purchased ? dayjs(detail.last_purchased).format("MMM D, YYYY") : "—"}
            />
          </nav>
        </div>

        <div class="column is-12">
          <svg id="d3-store-price-timeline" height="300" width="100%"></svg>
        </div>

        <div class="column is-12">
          <p class="is-size-6 has-text-weight-semibold has-text-grey mb-3">Items Purchased Here</p>

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
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><circle cx="11" cy="11" r="8" /><line
                      x1="21"
                      y1="21"
                      x2="16.65"
                      y2="16.65"
                    /></svg
                  >
                </span>
              </div>
            </div>
          </div>

          <div class="box has-background-white p-0" style="overflow-x: auto;">
            <table class="table is-fullwidth is-hoverable mb-0">
              <thead>
                <tr>
                  <th class="is-clickable" on:click={() => toggleSort("name")}>
                    Item {sortBy === "name" ? (sortAsc ? "↑" : "↓") : ""}
                  </th>
                  <th class="has-text-right is-clickable" on:click={() => toggleSort("latest_price")}>
                    Latest Price {sortBy === "latest_price" ? (sortAsc ? "↑" : "↓") : ""}
                  </th>
                  <th class="has-text-right">vs Other Stores</th>
                  <th class="has-text-right is-clickable" on:click={() => toggleSort("change")}>
                    30d Change {sortBy === "change" ? (sortAsc ? "↑" : "↓") : ""}
                  </th>
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
                      {formatCurrency(item.latestPrice)}/{item.unit}
                    </td>
                    <td class="has-text-right">
                      {#if item.vsOtherStores !== null}
                        <span style="color: {changeColor(item.vsOtherStores)}">
                          {formatChange(item.vsOtherStores)}
                        </span>
                      {:else}
                        <span class="has-text-grey">only store</span>
                      {/if}
                    </td>
                    <td class="has-text-right">
                      {#if item.change !== 0}
                        <span style="color: {changeColor(item.change)}"
                          >{formatChange(item.change)}</span
                        >
                      {:else}
                        <span class="has-text-grey">—</span>
                      {/if}
                    </td>
                    <td class="has-text-right">{dayjs(item.lastPurchased).format("MMM D, YYYY")}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          {#if filteredItems.length === 0 && searchQuery}
            <p class="has-text-grey has-text-centered mt-4">No items matching "{searchQuery}"</p>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</section>
```

- [ ] **Step 2: Verify frontend compiles**

Run: `npx svelte-check --threshold warning 2>&1 | head -30`
Expected: No new errors

- [ ] **Step 3: Commit**

```bash
git add src/routes/\(app\)/more/price_tracking/store/\[name\]/+page.svelte
git commit -m "feat(price-tracking): add store detail page with chart and items table"
```

---

### Task 5: Add back button to item detail page

**Files:**
- Modify: `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte:153-155`

- [ ] **Step 1: Add back button**

In `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`, find this block (lines 153-155):

```svelte
        <div class="column is-12">
          <h2 class="title is-4 mb-2">{detail.name}</h2>
```

Replace with:

```svelte
        <div class="column is-12">
          <button
            class="button is-small is-ghost px-0 mb-3"
            style="text-decoration: none;"
            on:click={() => history.back()}
          >
            ← Back
          </button>
          <h2 class="title is-4 mb-2">{detail.name}</h2>
```

- [ ] **Step 2: Commit**

```bash
git add src/routes/\(app\)/more/price_tracking/item/\[name\]/+page.svelte
git commit -m "feat(price-tracking): add contextual back button to item detail page"
```

---

### Task 6: Make store names clickable links

**Files:**
- Modify: `src/routes/(app)/more/price_tracking/+page.svelte:141-145` (Store Rankings)
- Modify: `src/routes/(app)/more/price_tracking/+page.svelte:151` (Recently Added)
- Modify: `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte:230` (Store Comparison table)

- [ ] **Step 1: Make store names clickable in Store Rankings**

In `src/routes/(app)/more/price_tracking/+page.svelte`, find (lines 141-145):

```svelte
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <span>
                  <span class="has-text-grey is-size-7 mr-2">#{i + 1}</span>
                  {ranking.store}
                </span>
```

Replace with:

```svelte
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <span>
                  <span class="has-text-grey is-size-7 mr-2">#{i + 1}</span>
                  <a class="secondary-link" href="/more/price_tracking/store/{ranking.store}">{ranking.store}</a>
                </span>
```

- [ ] **Step 2: Make store names clickable in Recently Added**

In the same file, find (line 153):

```svelte
                <span>{entry.store}</span>
```

Replace with:

```svelte
                <a class="secondary-link" href="/more/price_tracking/store/{entry.store}">{entry.store}</a>
```

- [ ] **Step 3: Make store names clickable in item detail Store Comparison table**

In `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`, find (line 230):

```svelte
                    <td>{sc.store}</td>
```

Replace with:

```svelte
                    <td><a class="secondary-link" href="/more/price_tracking/store/{sc.store}">{sc.store}</a></td>
```

- [ ] **Step 4: Commit**

```bash
git add src/routes/\(app\)/more/price_tracking/+page.svelte src/routes/\(app\)/more/price_tracking/item/\[name\]/+page.svelte
git commit -m "feat(price-tracking): make store names clickable links to store detail"
```

---

### Task 7: Manual verification

- [ ] **Step 1: Start dev server**

Run: `make develop`

- [ ] **Step 2: Verify store detail page**

1. Navigate to `/more/price_tracking` in the browser
2. Click a store name in Store Rankings — should navigate to `/more/price_tracking/store/{name}`
3. Verify KPI metrics display correctly (Items Tracked, Avg Unit Price, Price Trend, Last Purchase)
4. Verify the D3 chart renders showing average unit price over time
5. Verify the items table shows all items with Latest Price, vs Other Stores comparison, 30d Change, Last Purchased
6. Verify search filtering works
7. Verify column sorting works (Item, Latest Price, 30d Change)
8. Verify item names link back to item detail pages

- [ ] **Step 3: Verify back buttons**

1. From dashboard, click an item → item detail page shows "← Back" button
2. Click "← Back" → returns to dashboard
3. From dashboard, click a store → store detail page shows "← Back" button
4. Click "← Back" → returns to dashboard
5. From item detail, click a store in comparison table → store detail
6. Click "← Back" → returns to item detail

- [ ] **Step 4: Verify store links everywhere**

1. Dashboard Store Rankings: store names are clickable links
2. Dashboard Recently Added: store names are clickable links
3. Item Detail Store Comparison table: store names are clickable links

- [ ] **Step 5: Run lint and tests**

Run: `make lint && make test`
Expected: All pass
