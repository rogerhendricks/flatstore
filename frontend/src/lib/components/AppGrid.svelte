<script lang="ts">
	import type { AppSummary } from '$lib/types';
	import AppCard from '$lib/components/AppCard.svelte';

	export let apps: AppSummary[];
	export let isLoading: boolean = false;
	export let errorMessage: string = '';
</script>

<div class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-6">
	{#if isLoading}
		<div class="col-span-full text-center text-muted-foreground py-10">Loading applications...</div>
	{:else if errorMessage}
		<div class="col-span-full text-center text-destructive py-10">Failed to load: {errorMessage}</div>
	{:else if apps.length === 0}
		<div class="col-span-full text-center text-muted-foreground py-10">No applications found.</div>
	{:else}
		{#each apps as app}
			<AppCard {app} />
		{/each}
	{/if}
</div>
