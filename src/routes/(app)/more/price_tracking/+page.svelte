<script lang="ts">
  import { ajax, formatCurrency, isMobile } from "$lib/utils";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import {
    formatChange,
    changeColor,
    type PriceTrackingDashboard,
    type PriceMover,
    type StoreRanking,
    type RecentEntry
  } from "$lib/price_tracking";
  import { onMount } from "svelte";
  import dayjs from "dayjs";

  let dashboard: PriceTrackingDashboard = null;
  let isEmpty = true;

  onMount(async () => {
    dashboard = await ajax("/api/price_tracking");
    isEmpty = dashboard.items_count === 0;
  });
</script>

<section class="section">
  <div class="container is-fluid">
    <div class="columns is-flex-wrap-wrap is-centered">
      <ZeroState item={!isEmpty}>
        <strong>No price data yet.</strong> Post receipt data to the
        <code>/api/price_tracking/receipt</code> endpoint to start tracking prices.
      </ZeroState>

      {#if dashboard && !isEmpty}
        <div class="column is-12">
          <nav class="level {isMobile() && 'grid-2'}">
            <LevelItem title="Items Tracked" value={String(dashboard.items_count)} />
            <LevelItem title="Stores" value={String(dashboard.stores_count)} />
            <LevelItem
              title="Avg Price Change (30d)"
              value={formatChange(dashboard.avg_change)}
              color={changeColor(dashboard.avg_change)}
            />
          </nav>
        </div>

        <div class="column is-6">
          <div class="box has-background-white">
            <p class="heading mb-3">Biggest Increases (30d)</p>
            {#each dashboard.increases as mover (mover.name)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <a class="secondary-link" href="/more/price_tracking/item/{mover.name}">
                  {mover.name} (1 {mover.unit})
                </a>
                <span style="color: {changeColor(mover.change)}">{formatChange(mover.change)}</span>
              </div>
            {/each}
            {#if dashboard.increases.length === 0}
              <p class="has-text-grey">Not enough data</p>
            {/if}
          </div>
        </div>

        <div class="column is-6">
          <div class="box has-background-white">
            <p class="heading mb-3">Biggest Decreases (30d)</p>
            {#each dashboard.decreases as mover (mover.name)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <a class="secondary-link" href="/more/price_tracking/item/{mover.name}">
                  {mover.name} (1 {mover.unit})
                </a>
                <span style="color: {changeColor(mover.change)}">{formatChange(mover.change)}</span>
              </div>
            {/each}
            {#if dashboard.decreases.length === 0}
              <p class="has-text-grey">Not enough data</p>
            {/if}
          </div>
        </div>

        <div class="column is-6">
          <div class="box has-background-white">
            <p class="heading mb-3">Cheapest Store by Item Count</p>
            {#each dashboard.store_ranking as ranking (ranking.store)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <span>{ranking.store}</span>
                <span class="has-text-weight-bold">{ranking.count} items cheapest</span>
              </div>
            {/each}
          </div>
        </div>

        <div class="column is-6">
          <div class="box has-background-white">
            <p class="heading mb-3">Recently Added</p>
            {#each dashboard.recent as entry (entry.date + entry.store)}
              <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                <span>{dayjs(entry.date).format("MMM D")}</span>
                <span>{entry.store}</span>
                <span>{entry.count} items</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  </div>
</section>

<style lang="scss">
  .border-bottom {
    border-bottom: 1px solid #eee;
    &:last-child {
      border-bottom: none;
    }
  }
</style>
