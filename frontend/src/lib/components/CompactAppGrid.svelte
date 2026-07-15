<script lang="ts">
	import type { AppSummary } from '$lib/types';
	import AppCardCompact from '$lib/components/AppCardCompact.svelte';

	export let apps: AppSummary[];

	// Distribute apps evenly across 3 columns (contiguous chunks), so each
	// column card has a uniform fixed-height row layout like the Discover
	// dashboard, regardless of how many apps are passed in.
	$: perCol = Math.ceil(apps.length / 3);
	$: columns = [0, 1, 2].map((i) => apps.slice(i * perCol, i * perCol + perCol));
</script>

{#if apps.length === 0}
	<p class="text-sm text-muted-foreground py-4">No applications found.</p>
{:else}
	<div class="overflow-x-auto">
		<div class="grid grid-cols-3 gap-3 min-w-[600px]">
			{#each columns as column}
				{#if column.length > 0}
					<div class="bg-card border border-border rounded-2xl overflow-hidden">
						{#each column as app, j}
							<AppCardCompact {app} bordered={j > 0} />
						{/each}
					</div>
				{/if}
			{/each}
		</div>
	</div>
{/if}
