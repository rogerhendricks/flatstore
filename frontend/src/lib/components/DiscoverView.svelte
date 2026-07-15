<script lang="ts">
	import { apps, isLoading, errorMessage, activeCategory, discoverShowAll, viewTitle } from '$lib/stores/appStore';
	import DiscoverDashboard from '$lib/components/DiscoverDashboard.svelte';
	import AppGrid from '$lib/components/AppGrid.svelte';
</script>

<section class="flex-1 overflow-y-auto">
	{#if $activeCategory === 'Discover' && !$discoverShowAll}
		<DiscoverDashboard />
	{:else}
		<!-- ─────────────────────────────────────────────────── -->
		<!-- STANDARD GRID  (categories / search / see-all)      -->
		<!-- ─────────────────────────────────────────────────── -->
		<div class="p-8">
			<header class="mb-8 flex items-center gap-3">
				{#if $activeCategory === 'Discover' && $discoverShowAll}
				<button
					class="text-sm font-medium text-primary hover:underline shrink-0"
					on:click={() => discoverShowAll.set(false)}
				>← Discover</button>
				<span class="text-muted-foreground">/</span>
				{/if}
				<h1 class="text-3xl font-bold tracking-tight">
					{$discoverShowAll ? 'New Apps and Updates' : $viewTitle}
				</h1>
			</header>

			<AppGrid apps={$apps} isLoading={$isLoading} errorMessage={$errorMessage} />
		</div>
	{/if}
</section>
