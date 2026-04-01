# Price Tracking Feature — Design Spec

## Overview

A standalone price tracking system for monitoring item prices over time across stores. Users post receipt data (preprocessed by AI) via an API endpoint, and the app provides a dashboard and item-level detail views showing price trends, store comparisons, and biggest movers.

This feature is **independent of the ledger/accounting system** — it's a separate price database with no connection to journal transactions.

## Data Model

**Table: `receipt_items`**

| Column     | Type            | Description                              |
|------------|-----------------|------------------------------------------|
| ID         | uint (PK)       | Auto-increment                           |
| Date       | time.Time       | Purchase date                            |
| Store      | string          | Store name                               |
| Name       | string          | Normalized product type (milk, eggs, gas) |
| Brand      | string          | Brand/manufacturer                       |
| Variant    | string          | Flavor, packaging, organic, etc.         |
| Unit       | string          | Unit of measure (gal, doz, lb)           |
| Quantity   | decimal.Decimal | Number of units purchased                |
| Price      | decimal.Decimal | Total line price paid                    |
| UnitPrice  | decimal.Decimal | Price / Quantity, computed at insert time |

**Go model:** `internal/model/receipt_item/receipt_item.go` — follows existing GORM model pattern with `AutoMigrate`, decimal fields via `github.com/shopspring/decimal`.

**Single flat table.** No separate receipts table. Each row is one item from one purchase at one store on one date.

## API

### POST `/api/price_tracking/receipt`

Accepts receipt data with multiple items. Computes `UnitPrice` server-side at insert time.

**Request:**
```json
{
  "store": "Costco",
  "date": "2026-03-28",
  "items": [
    {
      "name": "milk",
      "brand": "Kirkland",
      "variant": "Whole, Organic",
      "unit": "gal",
      "quantity": 2,
      "price": 9.78
    },
    {
      "name": "eggs",
      "brand": "Happy Egg",
      "variant": "Free Range, Large",
      "unit": "doz",
      "quantity": 1,
      "price": 5.49
    }
  ]
}
```

**Response:**
```json
{"success": true, "count": 2}
```

**Validation:** `store`, `date`, and at least one item required. Each item requires `name`, `unit`, `quantity > 0`, `price >= 0`.

### GET `/api/price_tracking`

Returns dashboard data:
- Summary stats: total items tracked, total stores, average price change (30d)
- Top 5 biggest price increases (30d) — item name, percentage change
- Top 5 biggest price decreases (30d) — item name, percentage change
- Cheapest store rankings — store name + count of items where it has the lowest latest unit price
- Recent entries — last 10 receipt submissions (date, store, item count)

### GET `/api/price_tracking/items`

Returns list of all distinct item names for search/browse.

### GET `/api/price_tracking/item/:name`

Returns full price history for a specific item:
- Latest unit price, 90-day average, 90-day percentage change
- All price entries (date, store, brand, variant, quantity, unit, unit price)
- Distinct stores, brands, and variants for that item (for filter dropdowns)

## Frontend

### Navigation

Add "Price Tracking" to the app sidebar/navigation, under `more/` or as a top-level route — following existing SvelteKit route conventions.

### Dashboard Page — `/price_tracking`

**Route:** `src/routes/(app)/more/price_tracking/+page.svelte`

Layout:
- **Top row:** 3 summary cards (items tracked, stores, avg 30d change)
- **Middle row (2 columns):**
  - Biggest increases card (top 5, item names link to detail page)
  - Biggest decreases card (top 5, item names link to detail page)
- **Bottom row (2 columns):**
  - Cheapest store rankings card
  - Recently added entries card

Styling: Uses existing DaisyUI card components and Tailwind grid layout to match Paisa's look and feel. Follows the pattern of other dashboard-style pages in the app.

### Item Detail Page — `/price_tracking/item/[name]`

**Route:** `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte`

Layout:
- **Header:** Item name, latest unit price, 90-day average, 90-day percentage change
- **Filters:** Dropdowns for store, brand, variant (populated from API response)
- **Chart:** D3 line chart showing unit price over time, one line per store (filterable)
- **Table:** Full history with columns: date, store, brand, variant, quantity, unit price, change %

### Frontend Module

**File:** `src/lib/price_tracking.ts`

Utility functions for:
- Data transformation and grouping
- D3 chart rendering (price timeline)
- Percentage change calculations for display

## Backend Files

| File | Purpose |
|------|---------|
| `internal/model/receipt_item/receipt_item.go` | GORM model + DB operations |
| `internal/server/price_tracking.go` | Gin handlers for all endpoints |

Handler functions follow existing pattern: `GetPriceTracking(db *gorm.DB) gin.H`, `GetPriceTrackingItem(db *gorm.DB, name string) gin.H`, `PostReceipt(db *gorm.DB, request ReceiptRequest) gin.H`.

Routes registered in `internal/server/server.go` alongside existing routes.

## Out of Scope

- No connection to ledger journal or existing accounting features
- No authentication on the POST endpoint beyond existing token auth middleware
- No item catalog or config — items are created implicitly on first POST
- No fuzzy matching or item normalization — AI preprocessing handles this upstream
- No receipt-level grouping or receipt history views
