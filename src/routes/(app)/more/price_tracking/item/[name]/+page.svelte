<script lang="ts">
  import { page } from "$app/stores";
  import { ajax, formatCurrency, isMobile, type Legend } from "$lib/utils";
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
  let chartLegends: Legend[] = [];

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
    await ajax(
      "/api/price_tracking/receipt_item/:id",
      {
        method: "PUT",
        body: JSON.stringify({
          store: editForm.store,
          brand: editForm.brand,
          variant: editForm.variant,
          quantity: parseFloat(editForm.quantity),
          price: parseFloat(editForm.price)
        })
      },
      { id: String(id) }
    );
    editingId = null;
    detail = await ajax("/api/price_tracking/item/:name", null, { name });
    setTimeout(renderChart, 0);
  }

  async function deleteEntry(id: number) {
    await ajax(
      "/api/price_tracking/receipt_item/:id",
      {
        method: "DELETE"
      },
      { id: String(id) }
    );
    deleteConfirmId = null;
    detail = await ajax("/api/price_tracking/item/:name", null, { name });
    setTimeout(renderChart, 0);
  }

  $: name = $page.params.name;

  $: filteredEntries = detail
    ? detail.entries.filter((e) => {
        if (filterStore && e.store !== filterStore) return false;
        if (filterBrand && e.brand !== filterBrand) return false;
        if (filterVariant && e.variant !== filterVariant) return false;
        return true;
      })
    : [];

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

  $: if (detail && filterStore !== undefined) {
    // Re-render chart when filters change (tick to ensure DOM is ready)
    setTimeout(renderChart, 0);
  }

  onMount(async () => {
    detail = await ajax("/api/price_tracking/item/:name", null, { name });
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

        {#if storeComparisons.length > 1}
          <div class="column is-12">
            {#if bestStore}
              <div class="notification is-light is-size-7 py-2 px-4 mb-3">
                Best price: <strong
                  >{formatCurrency(bestStore.latestPrice)}/{detail.latest.unit}</strong
                >
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
                    <td><a class="secondary-link" href="/more/price_tracking/store/{sc.store}">{sc.store}</a></td>
                    <td class="has-text-right"
                      >{formatCurrency(sc.latestPrice)}/{detail.latest.unit}</td
                    >
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

        <div class="column is-12">
          <svg id="d3-price-timeline" height="300" width="100%"></svg>
          {#if chartLegends.length > 1}
            <div class="is-flex is-flex-wrap-wrap mt-2" style="gap: 12px;">
              {#each chartLegends as legend}
                <button
                  class="button is-small is-ghost px-2"
                  style="text-decoration: none;"
                  on:click={() => {
                    if (legend.toggle) legend.toggle();
                  }}
                >
                  <span
                    style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: {legend.color}; margin-right: 6px;"
                  ></span>
                  {legend.label}
                </button>
              {/each}
            </div>
          {/if}
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
                  <th class="has-text-right">Change / Actions</th>
                </tr>
              </thead>
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
                        <button
                          class="button is-small is-success mr-1"
                          on:click={() => saveEdit(entry.id)}>Save</button
                        >
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
                          <button
                            class="button is-small is-danger mr-1"
                            on:click={() => deleteEntry(entry.id)}>Yes</button
                          >
                          <button class="button is-small" on:click={() => (deleteConfirmId = null)}
                            >No</button
                          >
                        {:else}
                          {#if pctChange !== null}
                            <span style="color: {changeColor(pctChange)}" class="mr-2"
                              >{formatChange(pctChange)}</span
                            >
                          {:else}
                            <span class="has-text-grey mr-2">—</span>
                          {/if}
                          <button
                            class="button is-small is-ghost px-1"
                            title="Edit"
                            on:click={() => startEdit(entry)}
                          >
                            ✎
                          </button>
                          <button
                            class="button is-small is-ghost px-1 has-text-danger"
                            title="Delete"
                            on:click={() => (deleteConfirmId = entry.id)}
                          >
                            ✕
                          </button>
                        {/if}
                      </td>
                    </tr>
                  {/if}
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {/if}
  </div>
</section>
