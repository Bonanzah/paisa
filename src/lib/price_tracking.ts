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

export interface ItemSummary {
  name: string;
  unit: string;
  latest_price: number;
  avg_price: number;
  change: number;
  stores_count: number;
  last_purchased: string;
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

  g.append("g")
    .attr("transform", `translate(0,${height})`)
    .call(d3.axisBottom(x).ticks(d3.timeMonth.every(1)).tickFormat(d3.timeFormat("%b %Y")));

  g.append("g").call(d3.axisLeft(y).tickFormat((d) => formatCurrency(d as number)));

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
        tooltip.style("display", "block").html(
          `<strong>${d.store}</strong><br/>` +
            `${formatCurrency(d.unit_price)}/${d.unit}<br/>` +
            `${new Date(d.date).toLocaleDateString("en-US", {
              month: "short",
              day: "numeric",
              year: "numeric"
            })}` +
            (d.variant ? `<br/>${d.variant}` : "")
        );
      })
      .on("mousemove", (event: MouseEvent) => {
        tooltip.style("left", event.pageX + 12 + "px").style("top", event.pageY - 10 + "px");
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
