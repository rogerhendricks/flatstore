<script lang="ts">
	import { Loader2 } from '@lucide/svelte';
	import { featuredPromos } from '$lib/constants';
	import { apps, isLoading, discoverShowAll } from '$lib/stores/appStore';
	import CompactAppGrid from '$lib/components/CompactAppGrid.svelte';
</script>

<!-- ─────────────────────────────────────────────────── -->
<!-- DISCOVER DASHBOARD                                   -->
<!-- ─────────────────────────────────────────────────── -->
<div class="px-8 pt-8 pb-10">
	<h1 class="text-3xl font-bold tracking-tight mb-6">Discover</h1>

	{#if $isLoading}
		<div class="flex items-center justify-center h-48 text-muted-foreground">
			<Loader2 class="w-5 h-5 animate-spin mr-2" />Loading...
		</div>
	{:else}
		<!-- ── Featured promo cards (placeholder; external API in future) ── -->
		<div class="grid grid-cols-3 gap-4 mb-8">
			{#each featuredPromos as promo}
			<div
				class="relative rounded-2xl overflow-hidden h-44 cursor-pointer group"
				style="background: linear-gradient(135deg, {promo.gradient.from}, {promo.gradient.to})"
			>
				<div class="absolute inset-0 p-5 flex flex-col justify-between text-white">
					<span class="text-xs font-bold uppercase tracking-widest" style="opacity:0.7">{promo.badge}</span>
					<div>
						<h3 class="text-xl font-bold leading-tight">{promo.title}</h3>
						<p class="text-sm mt-1" style="opacity:0.75">{promo.subtitle}</p>
					</div>
				</div>
				<div class="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors rounded-2xl"></div>
			</div>
			{/each}
		</div>

		<!-- ── Section header ── -->
		<hr class="border-border mb-5" />
		<div class="flex items-center justify-between mb-2">
			<h2 class="text-lg font-bold tracking-tight">New Apps and Updates</h2>
			<button
				class="text-sm font-medium text-primary hover:underline"
				on:click={() => discoverShowAll.set(true)}
			>See All</button>
		</div>

		<!-- ── 3 × N app columns (uniform fixed-height cells) ── -->
		<CompactAppGrid apps={$apps} />
	{/if}
</div>
