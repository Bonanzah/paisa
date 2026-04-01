# Price Tracking v2 — Design Spec

## Overview

Improve the existing price tracking feature across four areas: visual polish (match Paisa's UI style), browsable items list, richer insights (store recommendations, comparisons), and manual data management (edit/delete entries). Also add server-side normalization to prevent data quality issues.

## 1. Dashboard Redesign

**Goal:** Replace generic card layout with Paisa-native styling. Make the page feel data-dense and consistent with the rest of the app.

**Layout:**

- **Tab bar** at the top of the page: "Dashboard" (default) and "Items" tabs using DaisyUI `du-tabs`.
- **Summary row:** 3 compact stat indicators in a horizontal Bulma `columns` row:
  - Items Tracked (count)
  - Stores (count)
  - Avg 30d Change (percentage, color-coded red/green)
- **Two-column grid** below the summary row using Bulma `columns`:
  - **Left column — Price Movers:** Single card with two sub-sections (Increases / Decreases). Each shows up to 5 items as compact rows: item name (link to detail), change %, latest unit price. Red for increases, green for decreases.
  - **Right column — Store Rankings + Recent:** Store rankings as a ranked list (store name, # items where cheapest). Below that, "Recently Added" as a compact table (date, store, item count). Last 10 entries.

**Style:** Use Paisa's existing `section`/`container` patterns, Bulma columns, muted gray/blue color palette. No custom card components — follow existing patterns from other Paisa pages.

## 2. Items List (New Tab)

**Goal:** Browsable, searchable list of all tracked items with at-a-glance stats.

**Layout (under "Items" tab):**

- **Search bar:** Text input at top, filters items by name in real-time (client-side).
- **Sort toggle:** Sort by name (A-Z default), latest price, or 30d change.
- **Table** with columns:
  - Name — clickable link to item detail page
  - Latest Price — unit price with unit label (e.g. "$4.89/gal")
  - 90d Avg — average unit price
  - 30d Change — percentage with trend arrow and color coding
  - Stores — count of distinct stores
  - Last Purchased — date of most recent entry

**Style:** Standard Bulma table matching Paisa's existing list views (transactions, holdings).

**Backend:** New endpoint `GET /api/price_tracking/items/summary` returning array of:
```json
{
  "name": "milk",
  "unit": "gal",
  "latest_price": 4.89,
  "avg_price": 4.65,
  "change": 5.2,
  "stores_count": 3,
  "last_purchased": "2026-03-28"
}
```

Computed server-side: for each distinct item name, query latest entry, calculate 90d average unit price, 30d change, count distinct stores, find max date.

## 3. Item Detail Page Improvements

### 3a. D3 Chart Polish

- **Tooltips:** On hover over a data point, show a tooltip with date, store, unit price, and variant.
- **Smooth lines:** Use D3 `curveMonotoneX` interpolation instead of straight line segments.
- **Legend:** Below the chart, show store names with their color dots. Clickable to toggle store visibility on the chart.
- **Y-axis:** Currency formatting with consistent decimal places.
- **X-axis:** Monthly date labels, not one per data point. Use D3 `timeMonth` ticks.

### 3b. Store Comparison & Recommendations (New Section)

Positioned below the filter bar, above the chart.

- **Best price banner:** Compact highlight: "Best price: **$3.49/gal** at **Costco** (saves 12% vs avg)"
  - Compares latest unit price per store, picks lowest.
  - "saves X%" is computed as `(avg_across_stores - best) / avg_across_stores * 100`.
- **Comparison table:** One row per store showing:
  - Store name
  - Latest unit price
  - vs-average % (positive = more expensive, negative = cheaper)
  - Last date seen at this store
  - Sorted cheapest first.

All data already available from the existing `GET /api/price_tracking/item/:name` response (entries array contains store, unit_price, date). Computation is client-side.

### 3c. Inline Edit & Delete

On the history table:

- **Edit:** Each row shows edit/delete icon buttons on hover (or always visible on mobile). Clicking edit converts the row into an inline form — text inputs for store, brand, variant, quantity, price. Date and name remain read-only. Save and Cancel buttons appear. On save, calls `PUT /api/price_tracking/receipt_item/:id`. Unit price is recomputed server-side.
- **Delete:** Clicking delete shows a confirmation prompt. On confirm, calls `DELETE /api/price_tracking/receipt_item/:id`. Row is removed from the table without a full page reload.

## 4. Data Normalization

**Applied server-side on all writes (POST receipt, PUT receipt_item):**

| Field   | Normalization          |
|---------|------------------------|
| name    | lowercase, trimmed     |
| store   | title case, trimmed    |
| brand   | title case, trimmed    |
| variant | lowercase, trimmed     |

**No retroactive migration** — existing seed data is already consistent. Users can fix inconsistencies via inline edit.

**No fuzzy matching** — "whole milk" and "milk" remain separate items. Keeps logic simple and predictable.

## 5. New API Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/price_tracking/items/summary` | Items list with computed stats |
| PUT | `/api/price_tracking/receipt_item/:id` | Update a receipt item entry |
| DELETE | `/api/price_tracking/receipt_item/:id` | Delete a receipt item entry |

### PUT `/api/price_tracking/receipt_item/:id`

**Request body:**
```json
{
  "store": "Costco",
  "brand": "Kirkland",
  "variant": "whole, organic",
  "quantity": 2,
  "price": 9.78
}
```

All fields optional — only provided fields are updated. `unit_price` is recomputed if quantity or price changes. Normalization rules applied to string fields.

**Response:** Updated `ReceiptItem` object.

**Errors:** 404 if ID not found. 400 if quantity <= 0 or price < 0.

### DELETE `/api/price_tracking/receipt_item/:id`

**Response:** `{"success": true}`

**Errors:** 404 if ID not found.

Hard delete (no soft delete — keeping it simple).

## 6. Files Changed

**Backend (Go):**
- `internal/server/price_tracking.go` — add PUT/DELETE handlers, items summary endpoint, normalization helper functions
- `internal/server/server.go` — register 3 new routes

**Frontend (Svelte/TS):**
- `src/routes/(app)/more/price_tracking/+page.svelte` — redesigned dashboard layout + items tab
- `src/routes/(app)/more/price_tracking/item/[name]/+page.svelte` — chart improvements, store comparison section, inline edit/delete
- `src/lib/price_tracking.ts` — new types (ItemSummary), updated renderPriceTimeline (tooltips, legend, smooth curves), normalization-aware formatting

**No changes to:**
- Database schema (no new columns or tables)
- Config files
- Navigation structure (routes stay the same)
- Any existing Paisa features
