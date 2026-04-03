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
  import { onMount, onDestroy } from "svelte";
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

  onDestroy(() => {
    if (chartDestroy) chartDestroy();
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
            <LevelItem title="Avg Unit Price" value={formatCurrency(detail.avg_unit_price)} />
            <LevelItem
              title="Price Trend (90d)"
              value={formatChange(detail.change)}
              color={changeColor(detail.change)}
            />
            <LevelItem
              title="Last Purchase"
              value={detail.last_purchased
                ? dayjs(detail.last_purchased).format("MMM D, YYYY")
                : "—"}
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
                  <th
                    class="has-text-right is-clickable"
                    on:click={() => toggleSort("latest_price")}
                  >
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
                    <td class="has-text-right">{dayjs(item.lastPurchased).format("MMM D, YYYY")}</td
                    >
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
