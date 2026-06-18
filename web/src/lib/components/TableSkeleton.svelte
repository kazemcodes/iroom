<!--
  TableSkeleton — Loading skeleton for data tables.
  
  Shows a shimmer animation that matches the table structure.
  Used while data is being fetched from the API.
  
  Props:
    rows: Number of skeleton rows to show (default: 5)
    cols: Number of skeleton columns (default: 4)
-->
<script lang="ts">
	let { rows = 5, cols = 4 }: { rows?: number; cols?: number } = $props();
</script>

<div class="skeleton-table">
	{#each Array(rows) as _, i}
		<div class="skeleton-row" style="animation-delay: {i * 0.05}s">
			{#each Array(cols) as _, j}
				<div class="skeleton-cell" style="width: {60 + Math.random() * 40}%; animation-delay: {j * 0.03}s"></div>
			{/each}
		</div>
	{/each}
</div>

<style>
	.skeleton-table {
		background: var(--color-pure);
		border-radius: 12px;
		padding: 16px;
		border: 1px solid var(--color-zen-garden);
	}
	.skeleton-row {
		display: flex;
		gap: 12px;
		padding: 12px 0;
		border-bottom: 1px solid var(--color-zen-garden);
	}
	.skeleton-row:last-child { border-bottom: none; }
	.skeleton-cell {
		height: 16px;
		border-radius: 4px;
		background: linear-gradient(90deg, var(--color-secret-glow) 25%, var(--color-dreamers-night) 50%, var(--color-secret-glow) 75%);
		background-size: 200% 100%;
		animation: shimmer 1.5s infinite;
	}
	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>
