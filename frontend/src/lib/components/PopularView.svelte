<script lang="ts">
	import { Sparkles, Gamepad2, Palette, LoaderCircle } from '@lucide/svelte';
	import { heroGradients } from '$lib/constants';
	import {
		popularTab,
		heroApps,
		restApps,
		heroIndex,
		isLoading,
		errorMessage,
		handleImageError,
		openDetails,
		loadPopular,
		startHeroSlide
	} from '$lib/stores/appStore';
	import CompactAppGrid from '$lib/components/CompactAppGrid.svelte';
</script>

<section class="flex-1 overflow-y-auto">
	<!-- Header + Tab Button Group -->
	<div class="px-8 pt-8 pb-6">
		<div class="inline-flex rounded-xl border border-border bg-muted p-1 gap-1">
			<button
				class="flex items-center gap-1.5 px-4 py-1.5 text-sm font-medium rounded-lg transition-colors
					{$popularTab === 'discover' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}"
				on:click={() => loadPopular('discover')}
			>
				<Sparkles class="w-3.5 h-3.5" />Discover
			</button>
			<button
				class="flex items-center gap-1.5 px-4 py-1.5 text-sm font-medium rounded-lg transition-colors
					{$popularTab === 'games' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}"
				on:click={() => loadPopular('games')}
			>
				<Gamepad2 class="w-3.5 h-3.5" />Games
			</button>
			<button
				class="flex items-center gap-1.5 px-4 py-1.5 text-sm font-medium rounded-lg transition-colors
					{$popularTab === 'create' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}"
				on:click={() => loadPopular('create')}
			>
				<Palette class="w-3.5 h-3.5" />Create
			</button>
		</div>
	</div>

	{#if $isLoading}
		<div class="flex items-center justify-center py-20 text-muted-foreground">
			<LoaderCircle class="w-5 h-5 animate-spin mr-2" />Loading...
		</div>
	{:else if $errorMessage}
		<div class="text-center text-destructive py-20">Failed to load: {$errorMessage}</div>
	{:else}
		<!-- Hero Carousel -->
		{#if $heroApps.length > 0}
		<div class="relative mx-8 mb-8 rounded-3xl overflow-hidden" style="height: 280px;">
			{#each $heroApps as app, i}
			{@const g = heroGradients[i % heroGradients.length]}
			<div
				class="absolute inset-0 flex items-center gap-8 p-8 transition-opacity duration-500
					{i === $heroIndex ? 'opacity-100 z-10' : 'opacity-0 z-0'}"
				style="background: linear-gradient(135deg, {g.from}, {g.via}, {g.to})"
			>
				<img
					class="w-28 h-28 rounded-2xl object-contain shrink-0 shadow-2xl cursor-pointer"
					src={app.iconUrl}
					alt={app.name}
					on:error={handleImageError}
					on:click={() => openDetails(app.flatpakAppId)}
				/>
				<div class="flex-1 text-white min-w-0 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>
					<p class="text-xs font-bold uppercase tracking-widest mb-2" style="opacity:0.6">
						#{i + 1} · Most Popular
					</p>
					<h2 class="text-3xl font-bold leading-tight truncate mb-1">{app.name}</h2>
					<p class="text-sm mb-3" style="opacity:0.75">{app.developer || 'Flathub'}</p>
					<p class="text-sm line-clamp-2 max-w-xl" style="opacity:0.65">{app.summary}</p>
				</div>
			</div>
			{/each}

			<!-- Pill dot navigation -->
			<div class="absolute bottom-5 left-0 right-0 flex justify-center gap-2 z-20">
				{#each $heroApps as _, i}
				<button
					class="rounded-full bg-white transition-all duration-200
						{i === $heroIndex ? 'w-6 h-2.5 opacity-100' : 'w-2.5 h-2.5 opacity-40 hover:opacity-70'}"
					on:click={() => { heroIndex.set(i); startHeroSlide(); }}
					aria-label="Slide {i + 1}"
				></button>
				{/each}
			</div>
		</div>
		{/if}

		<!-- Rest of apps grid -->
		{#if $restApps.length > 0}
		<div class="px-8 pb-8">
			<h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-4">More Apps</h2>
			<CompactAppGrid apps={$restApps} />
		</div>
		{/if}
	{/if}
</section>
