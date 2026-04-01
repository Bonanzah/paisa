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
