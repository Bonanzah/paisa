<script lang="ts">
  import { ajax, isMobile } from "$lib/utils";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import { formatChange, changeColor, type PriceTrackingDashboard } from "$lib/price_tracking";
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
    <ZeroState item={!isEmpty}>
      <strong>No price data yet.</strong> Post receipt data to the
      <code>/api/price_tracking/receipt</code> endpoint to start tracking prices.
    </ZeroState>

    {#if dashboard && !isEmpty}
      <div class="columns is-multiline">
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
          <div class="content">
            <p class="subtitle has-text-grey">Biggest Increases (30d)</p>
            <div class="box px-3 has-background-white">
              {#each dashboard.increases ?? [] as mover (mover.name)}
                <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                  <a class="secondary-link" href="/more/price_tracking/item/{mover.name}">
                    {mover.name} (1 {mover.unit})
                  </a>
                  <span class="has-text-weight-bold" style="color: {changeColor(mover.change)}"
                    >{formatChange(mover.change)}</span
                  >
                </div>
              {/each}
              {#if (dashboard.increases ?? []).length === 0}
                <p class="has-text-grey">Not enough data</p>
              {/if}
            </div>
          </div>
        </div>

        <div class="column is-6">
          <div class="content">
            <p class="subtitle has-text-grey">Biggest Decreases (30d)</p>
            <div class="box px-3 has-background-white">
              {#each dashboard.decreases ?? [] as mover (mover.name)}
                <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                  <a class="secondary-link" href="/more/price_tracking/item/{mover.name}">
                    {mover.name} (1 {mover.unit})
                  </a>
                  <span class="has-text-weight-bold" style="color: {changeColor(mover.change)}"
                    >{formatChange(mover.change)}</span
                  >
                </div>
              {/each}
              {#if (dashboard.decreases ?? []).length === 0}
                <p class="has-text-grey">Not enough data</p>
              {/if}
            </div>
          </div>
        </div>

        <div class="column is-6">
          <div class="content">
            <p class="subtitle has-text-grey">Cheapest Store by Item Count</p>
            <div class="box px-3 has-background-white">
              {#each dashboard.store_ranking ?? [] as ranking (ranking.store)}
                <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                  <span>{ranking.store}</span>
                  <span class="has-text-weight-bold">{ranking.count} items cheapest</span>
                </div>
              {/each}
            </div>
          </div>
        </div>

        <div class="column is-6">
          <div class="content">
            <p class="subtitle has-text-grey">Recently Added</p>
            <div class="box px-3 has-background-white">
              {#each dashboard.recent ?? [] as entry (entry.date + entry.store)}
                <div class="is-flex is-justify-content-space-between py-2 border-bottom">
                  <span>{dayjs(entry.date).format("MMM D")}</span>
                  <span>{entry.store}</span>
                  <span>{entry.count} items</span>
                </div>
              {/each}
            </div>
          </div>
        </div>
      </div>
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
