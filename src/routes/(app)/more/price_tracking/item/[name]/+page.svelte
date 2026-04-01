<script lang="ts">
  import { page } from "$app/stores";
  import { ajax, formatCurrency, isMobile } from "$lib/utils";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import {
    formatChange,
    changeColor,
    renderPriceTimeline,
    type PriceTrackingItemDetail
  } from "$lib/price_tracking";
  import { onMount } from "svelte";
  import dayjs from "dayjs";

  let detail: PriceTrackingItemDetail = null;
  let filterStore = "";
  let filterBrand = "";
  let filterVariant = "";
  let chartDestroy: () => void = null;

  $: name = $page.params.name;

  $: filteredEntries = detail
    ? detail.entries.filter((e) => {
        if (filterStore && e.store !== filterStore) return false;
        if (filterBrand && e.brand !== filterBrand) return false;
        if (filterVariant && e.variant !== filterVariant) return false;
        return true;
      })
    : [];

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

        <div class="column is-12">
          <svg id="d3-price-timeline" height="300" width="100%"></svg>
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
                  <th class="has-text-right">Change</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredEntries as entry, i (entry.id)}
                  {@const prev = filteredEntries[i + 1]}
                  {@const pctChange =
                    prev && prev.unit_price !== 0
                      ? ((entry.unit_price - prev.unit_price) / prev.unit_price) * 100
                      : null}
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
                      {#if pctChange !== null}
                        <span style="color: {changeColor(pctChange)}"
                          >{formatChange(pctChange)}</span
                        >
                      {:else}
                        <span class="has-text-grey">—</span>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {/if}
  </div>
</section>
