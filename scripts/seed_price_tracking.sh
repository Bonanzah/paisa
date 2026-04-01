#!/bin/bash
# Seed price tracking with sample receipt data across multiple stores and dates.
# Usage: ./scripts/seed_price_tracking.sh [BASE_URL]

BASE_URL="${1:-http://localhost:5173}"
ENDPOINT="$BASE_URL/api/price_tracking/receipt"

post() {
  curl -s -X POST "$ENDPOINT" -H "Content-Type: application/json" -d "$1"
  echo ""
}

# --- Walmart, Feb 1 ---
post '{
  "store": "Walmart",
  "date": "2026-02-01",
  "items": [
    {"name": "milk", "brand": "Great Value", "variant": "Whole", "unit": "gal", "quantity": 1, "price": 3.98},
    {"name": "eggs", "brand": "Great Value", "variant": "Large, 12ct", "unit": "doz", "quantity": 1, "price": 4.12},
    {"name": "bread", "brand": "Sara Lee", "variant": "White", "unit": "loaf", "quantity": 1, "price": 3.49},
    {"name": "chicken breast", "brand": "Tyson", "variant": "Boneless Skinless", "unit": "lb", "quantity": 3, "price": 11.97},
    {"name": "rice", "brand": "Great Value", "variant": "Long Grain White", "unit": "lb", "quantity": 5, "price": 4.48},
    {"name": "bananas", "brand": "", "variant": "Yellow", "unit": "lb", "quantity": 2, "price": 1.14},
    {"name": "gas", "brand": "Murphy USA", "variant": "Regular 87", "unit": "gal", "quantity": 12, "price": 37.08}
  ]
}'

# --- Costco, Feb 5 ---
post '{
  "store": "Costco",
  "date": "2026-02-05",
  "items": [
    {"name": "milk", "brand": "Kirkland", "variant": "Whole, Organic", "unit": "gal", "quantity": 2, "price": 9.58},
    {"name": "eggs", "brand": "Kirkland", "variant": "Large, 24ct", "unit": "doz", "quantity": 2, "price": 8.98},
    {"name": "chicken breast", "brand": "Kirkland", "variant": "Boneless Skinless", "unit": "lb", "quantity": 6, "price": 20.94},
    {"name": "rice", "brand": "Kirkland", "variant": "Jasmine", "unit": "lb", "quantity": 25, "price": 18.99},
    {"name": "butter", "brand": "Kirkland", "variant": "Unsalted", "unit": "lb", "quantity": 4, "price": 12.49},
    {"name": "olive oil", "brand": "Kirkland", "variant": "Extra Virgin", "unit": "oz", "quantity": 67.6, "price": 16.99}
  ]
}'

# --- Trader Joes, Feb 10 ---
post '{
  "store": "Trader Joes",
  "date": "2026-02-10",
  "items": [
    {"name": "milk", "brand": "Trader Joes", "variant": "Whole, Organic", "unit": "gal", "quantity": 1, "price": 5.29},
    {"name": "eggs", "brand": "Trader Joes", "variant": "Free Range, Large", "unit": "doz", "quantity": 1, "price": 4.99},
    {"name": "bread", "brand": "Trader Joes", "variant": "Sourdough", "unit": "loaf", "quantity": 1, "price": 4.49},
    {"name": "bananas", "brand": "", "variant": "Yellow", "unit": "lb", "quantity": 1.5, "price": 0.88},
    {"name": "butter", "brand": "Trader Joes", "variant": "Cultured Salted", "unit": "lb", "quantity": 1, "price": 3.99},
    {"name": "olive oil", "brand": "Trader Joes", "variant": "Extra Virgin", "unit": "oz", "quantity": 33.8, "price": 8.99}
  ]
}'

# --- Walmart, Feb 20 ---
post '{
  "store": "Walmart",
  "date": "2026-02-20",
  "items": [
    {"name": "milk", "brand": "Great Value", "variant": "Whole", "unit": "gal", "quantity": 1, "price": 4.12},
    {"name": "eggs", "brand": "Great Value", "variant": "Large, 12ct", "unit": "doz", "quantity": 1, "price": 4.48},
    {"name": "bread", "brand": "Sara Lee", "variant": "White", "unit": "loaf", "quantity": 1, "price": 3.49},
    {"name": "chicken breast", "brand": "Tyson", "variant": "Boneless Skinless", "unit": "lb", "quantity": 2, "price": 8.38},
    {"name": "bananas", "brand": "", "variant": "Yellow", "unit": "lb", "quantity": 2.5, "price": 1.48},
    {"name": "gas", "brand": "Murphy USA", "variant": "Regular 87", "unit": "gal", "quantity": 10, "price": 32.90}
  ]
}'

# --- Costco, Mar 1 ---
post '{
  "store": "Costco",
  "date": "2026-03-01",
  "items": [
    {"name": "milk", "brand": "Kirkland", "variant": "Whole, Organic", "unit": "gal", "quantity": 2, "price": 9.78},
    {"name": "eggs", "brand": "Kirkland", "variant": "Large, 24ct", "unit": "doz", "quantity": 2, "price": 9.48},
    {"name": "chicken breast", "brand": "Kirkland", "variant": "Boneless Skinless", "unit": "lb", "quantity": 6, "price": 22.14},
    {"name": "butter", "brand": "Kirkland", "variant": "Unsalted", "unit": "lb", "quantity": 4, "price": 13.49},
    {"name": "olive oil", "brand": "Kirkland", "variant": "Extra Virgin", "unit": "oz", "quantity": 67.6, "price": 17.49},
    {"name": "gas", "brand": "Costco", "variant": "Regular 87", "unit": "gal", "quantity": 14, "price": 41.86}
  ]
}'

# --- Trader Joes, Mar 10 ---
post '{
  "store": "Trader Joes",
  "date": "2026-03-10",
  "items": [
    {"name": "milk", "brand": "Trader Joes", "variant": "Whole, Organic", "unit": "gal", "quantity": 1, "price": 5.49},
    {"name": "eggs", "brand": "Trader Joes", "variant": "Free Range, Large", "unit": "doz", "quantity": 1, "price": 5.49},
    {"name": "bread", "brand": "Trader Joes", "variant": "Sourdough", "unit": "loaf", "quantity": 1, "price": 4.49},
    {"name": "bananas", "brand": "", "variant": "Yellow", "unit": "lb", "quantity": 2, "price": 1.18},
    {"name": "butter", "brand": "Trader Joes", "variant": "Cultured Salted", "unit": "lb", "quantity": 1, "price": 4.29}
  ]
}'

# --- Walmart, Mar 15 ---
post '{
  "store": "Walmart",
  "date": "2026-03-15",
  "items": [
    {"name": "milk", "brand": "Great Value", "variant": "Whole", "unit": "gal", "quantity": 1, "price": 4.28},
    {"name": "eggs", "brand": "Great Value", "variant": "Large, 12ct", "unit": "doz", "quantity": 2, "price": 9.38},
    {"name": "bread", "brand": "Sara Lee", "variant": "White", "unit": "loaf", "quantity": 1, "price": 3.69},
    {"name": "chicken breast", "brand": "Tyson", "variant": "Boneless Skinless", "unit": "lb", "quantity": 3, "price": 13.47},
    {"name": "rice", "brand": "Great Value", "variant": "Long Grain White", "unit": "lb", "quantity": 5, "price": 4.68},
    {"name": "bananas", "brand": "", "variant": "Yellow", "unit": "lb", "quantity": 2, "price": 1.24},
    {"name": "gas", "brand": "Murphy USA", "variant": "Regular 87", "unit": "gal", "quantity": 11, "price": 36.63}
  ]
}'

# --- Costco, Mar 28 ---
post '{
  "store": "Costco",
  "date": "2026-03-28",
  "items": [
    {"name": "milk", "brand": "Kirkland", "variant": "Whole, Organic", "unit": "gal", "quantity": 2, "price": 9.98},
    {"name": "eggs", "brand": "Kirkland", "variant": "Large, 24ct", "unit": "doz", "quantity": 2, "price": 10.48},
    {"name": "chicken breast", "brand": "Kirkland", "variant": "Boneless Skinless", "unit": "lb", "quantity": 6, "price": 23.94},
    {"name": "rice", "brand": "Kirkland", "variant": "Jasmine", "unit": "lb", "quantity": 25, "price": 19.49},
    {"name": "butter", "brand": "Kirkland", "variant": "Unsalted", "unit": "lb", "quantity": 4, "price": 14.49},
    {"name": "olive oil", "brand": "Kirkland", "variant": "Extra Virgin", "unit": "oz", "quantity": 67.6, "price": 18.49},
    {"name": "gas", "brand": "Costco", "variant": "Regular 87", "unit": "gal", "quantity": 13, "price": 39.91}
  ]
}'

echo ""
echo "Done! Seeded 8 receipts across 3 stores (Feb-Mar 2026)."
echo "View dashboard at: $BASE_URL (navigate to More > Price Tracking)"
