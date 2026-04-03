# Store Detail Page & Back Button Navigation

## Overview

Add a store detail page to the price tracking feature that shows price trends and items purchased at a specific store. Add contextual back buttons to all detail pages.

## Store Detail Page

**Route:** `/more/price_tracking/store/[name]/+page.svelte`

### KPI Metrics Row

Four metrics displayed using the existing `LevelItem` component:

- **Items Tracked** — count of distinct item names purchased at this store
- **Avg Unit Price** — average unit price across all items at this store
- **Price Trend (90d)** — percentage change in average unit price comparing latest vs 90 days ago
- **Last Purchase** — most recent purchase date at this store

### Average Unit Price Chart

D3 line chart showing average unit price over time at this store. Reuses the rendering pattern from `renderPriceTimeline` in `src/lib/price_tracking.ts` but adapted for store-level data:

- X-axis: time (monthly ticks)
- Y-axis: average unit price across all items purchased on each date
- Single line (one store), no legend needed
- Same tooltip style as item detail chart

### Items Table

Table of all items purchased at this store with search and sortable columns:

| Column | Description |
|--------|-------------|
| Item | Item name, links to `/more/price_tracking/item/{name}` |
| Latest Price | Most recent unit price at this store for this item |
| vs Other Stores | Percentage difference from average price across all stores for same item. Shows "only store" if item only exists at this store |
| 30d Change | Price change at this store over last 30 days |
| Last Purchased | Most recent purchase date for this item at this store |

Search bar filters by item name. Columns sortable: Item (alpha), Latest Price, 30d Change.

## Backend API

**New endpoint:** `GET /api/price_tracking/store/:name`

**New function:** `GetPriceTrackingStore(db *gorm.DB, name string) gin.H`

**Response shape:**
```json
{
  "store": "Woolworths",
  "entries": [...],
  "items": ["milk", "bread", ...],
  "item_count": 24,
  "avg_unit_price": 4.82,
  "change": 3.2,
  "last_purchased": "2026-03-28T00:00:00Z"
}
```

- `entries` — all `ReceiptItem` records for this store, ordered by date DESC (used for chart rendering)
- `items` — distinct item names at this store
- `item_count` — count of distinct items
- `avg_unit_price` — average unit price across all entries
- `change` — 90-day price trend (same calculation pattern as item detail)
- `last_purchased` — most recent entry date

**New model function:** `ByStore(db *gorm.DB, store string) []ReceiptItem` in `internal/model/receipt_item/receipt_item.go`

The "vs Other Stores" comparison is computed client-side: for each item at this store, fetch the global items summary (already available via `/api/price_tracking/items/summary`) and compare.

## Frontend Types

Add to `src/lib/price_tracking.ts`:

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

Add a new chart function `renderStoreAvgPriceTimeline(selector, entries)` that computes per-date average unit price from entries and renders a single-line D3 chart.

## Back Button

Add a contextual back button to both detail pages:

- **Store detail page** — `← Back` link at top, calls `history.back()`
- **Item detail page** — same `← Back` link added at top of existing page

Implementation: a simple `<a>` or `<button>` styled consistently, using `on:click={() => history.back()}`.

## Navigation Links (Store Names Become Clickable)

Two locations get store name links:

1. **Dashboard Store Rankings** — wrap `{ranking.store}` in `<a href="/more/price_tracking/store/{ranking.store}">`
2. **Item Detail Store Comparison Table** — wrap `{sc.store}` in `<a href="/more/price_tracking/store/{sc.store}">`

## File Changes Summary

| File | Change |
|------|--------|
| `internal/model/receipt_item/receipt_item.go` | Add `ByStore()` query |
| `internal/server/price_tracking.go` | Add `GetPriceTrackingStore()` function |
| `internal/server/server.go` | Register `GET /api/price_tracking/store/:name` |
| `src/lib/price_tracking.ts` | Add `PriceTrackingStoreDetail` type, `renderStoreAvgPriceTimeline()` |
| `src/routes/(app)/more/price_tracking/store/[name]/+page.svelte` | New store detail page |
| `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte` | Add back button |
| `src/routes/(app)/more/price_tracking/+page.svelte` | Make store names clickable in Store Rankings and Recently Added |
