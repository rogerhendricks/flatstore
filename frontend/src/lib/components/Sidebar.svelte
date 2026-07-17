<script lang="ts">
	import {
		Sparkles,
		TrendingUp,
		HardDriveDownload,
		Sun,
		Moon,
		Laptop,
		SunMoon,
		CloudBackup,
		PackageCheck,
		Check,
		LoaderCircle,
		CircleAlert
	} from '@lucide/svelte';
	import * as Popover from '$lib/components/ui/popover';
	import { Progress } from '$lib/components/ui/progress';
	import { Badge } from '$lib/components/ui/badge'; // New import
	import { categories } from '$lib/constants';
	import {
		searchQuery,
		activeCategory,
		appProgress,
		activeTasksCount,
		updateableAppsCount, // New import
		currentTheme,
		applyTheme,
		getProgressColorClass,
		loadDiscover,
		loadCategory,
		loadInstalled,
		loadPopular,
		handleSearch,
		catalogStatus,
		catalogError
	} from '$lib/stores/appStore';
</script>

<aside class="flex flex-col w-52 min-w-[13rem] bg-card border-r border-border p-3 overflow-y-auto">
	<input
		type="text"
		class="w-full px-3 py-1.5 mb-4 text-sm bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
		placeholder="Search apps..."
		spellcheck="false"
		bind:value={$searchQuery}
		on:input={handleSearch}
	/>

	{#if $catalogStatus === 'syncing'}
		<div class="flex items-center gap-1.5 px-2 mb-3 text-[11px] text-muted-foreground" title="The AppStream catalog is downloading in the background. App details may show limited info until it finishes.">
			<LoaderCircle class="w-3 h-3 animate-spin shrink-0" />
			<span class="truncate">Syncing app catalog...</span>
		</div>
	{:else if $catalogStatus === 'error'}
		<div class="flex items-center gap-1.5 px-2 mb-3 text-[11px] text-amber-600 dark:text-amber-400" title={$catalogError}>
			<CircleAlert class="w-3 h-3 shrink-0" />
			<span class="truncate">Catalog sync failed</span>
		</div>
	{/if}

	<h3 class="text-xs font-semibold tracking-wider text-muted-foreground uppercase mb-2 px-2">
		Categories
	</h3>

	<ul class="flex-1 space-y-0.5">
		<li>
			<button
				class="w-full flex items-center gap-2 px-2 py-2 text-sm font-medium rounded-lg transition-colors
                       {$activeCategory === 'Discover'
					? 'bg-primary/10 text-primary'
					: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
				on:click={loadDiscover}
			><Sparkles class="w-4 h-4 shrink-0" />Discover</button>
		</li>
		<li>
			<button
				class="w-full flex items-center gap-2 px-2 py-2 text-sm font-medium rounded-lg transition-colors
					{$activeCategory === 'Popular' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
				on:click={() => loadPopular('discover')}
			>
				<TrendingUp class="w-4 h-4 shrink-0" />
				Popular
			</button>
		</li>
		{#each categories as cat}
			<li>
				<button
					class="w-full flex items-center gap-2 px-2 py-2 text-sm font-medium rounded-lg transition-colors
                           {$activeCategory === cat.id
						? 'bg-primary/10 text-primary'
						: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
					on:click={() => loadCategory(cat.id, cat.label)}
				>
					<svelte:component this={cat.icon} class="w-4 h-4 shrink-0" />
					{cat.label}
				</button>
			</li>
		{/each}
		<li>
			<button
				class="w-full flex items-center gap-2 px-2 py-2 text-sm font-medium rounded-lg transition-colors
					{$activeCategory === 'Installed' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
				on:click={loadInstalled}
			>
				<HardDriveDownload class="w-4 h-4 shrink-0" />
				Installed Apps
				{#if $updateableAppsCount > 0}
					<Badge variant="destructive" class="ml-auto px-2 py-0.5 text-xs">{$updateableAppsCount}</Badge>
				{/if}
			</button>
		</li>
	</ul>

	<div class="pt-3 mt-3 border-t border-border flex items-center">
		<Popover.Root>
			<Popover.Trigger
				class="relative p-2 rounded-xl hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
				title="Theme"
			>
				<SunMoon class="w-5 h-5" />
			</Popover.Trigger>

			<Popover.Content side="right" align="end" class="w-64 p-3">
				<!-- Theme -->
				<p class="text-xs font-semibold tracking-wider text-muted-foreground uppercase px-2 mb-2">Theme</p>
				<div class="flex flex-col gap-0.5">
					<button
						class="flex items-center gap-3 px-2 py-2 rounded-lg text-sm w-full transition-colors
                            {$currentTheme === 'light' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
						on:click={() => applyTheme('light')}
					>
						<Sun class="w-4 h-4 shrink-0" />
						<span>Light</span>
						{#if $currentTheme === 'light'}<Check class="w-4 h-4 ml-auto" />{/if}
					</button>
					<button
						class="flex items-center gap-3 px-2 py-2 rounded-lg text-sm w-full transition-colors
                            {$currentTheme === 'dark' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
						on:click={() => applyTheme('dark')}
					>
						<Moon class="w-4 h-4 shrink-0" />
						<span>Dark</span>
						{#if $currentTheme === 'dark'}<Check class="w-4 h-4 ml-auto" />{/if}
					</button>
					<button
						class="flex items-center gap-3 px-2 py-2 rounded-lg text-sm w-full transition-colors
                            {$currentTheme === 'system' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
						on:click={() => applyTheme('system')}
					>
						<Laptop class="w-4 h-4 shrink-0" />
						<span>System</span>
						{#if $currentTheme === 'system'}<Check class="w-4 h-4 ml-auto" />{/if}
					</button>
				</div>
			</Popover.Content>
		</Popover.Root>
		<Popover.Root>
			<Popover.Trigger
				class="relative p-2 rounded-xl hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
				title="Backup/ Restore"
			>
				<CloudBackup class="w-5 h-5" />
			</Popover.Trigger>

			<Popover.Content side="right" align="end" class="w-64 p-3">
				<p class="text-xs font-semibold tracking-wider text-muted-foreground uppercase px-2 mb-2">Backup/ Restore</p>
				<div class="flex flex-col gap-0.5"></div>
			</Popover.Content>
		</Popover.Root>
		<Popover.Root>
			<Popover.Trigger
				class="relative p-2 rounded-xl hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
				title="Activity Center"
			>
				<PackageCheck class="w-5 h-5" />
				{#if $activeTasksCount > 0}
					<span class="absolute top-1 right-1 flex h-3 w-3">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
						<span class="relative inline-flex rounded-full h-3 w-3 bg-blue-500"></span>
					</span>
				{/if}
			</Popover.Trigger>

			<Popover.Content side="right" align="end" class="w-80 p-4">
				<h4 class="font-medium text-sm mb-4">Background Tasks</h4>
				{#if Object.keys($appProgress).length === 0}
					<p class="text-sm text-muted-foreground text-center py-4">No active tasks.</p>
				{:else}
					<div class="space-y-4">
						{#each Object.entries($appProgress) as [id, prog]}
							<div class="space-y-2">
								<div class="flex justify-between text-xs font-medium">
									<span class="truncate pr-2">{prog.name}</span>
									<span class="capitalize text-muted-foreground">{prog.status} {prog.percentage}%</span>
								</div>
								<Progress value={prog.percentage} class="h-2 {getProgressColorClass(prog.status)}" />
							</div>
						{/each}
					</div>
				{/if}
			</Popover.Content>
		</Popover.Root>
	</div>
</aside>
