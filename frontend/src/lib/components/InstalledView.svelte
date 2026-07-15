<script lang="ts">
	import { Loader2, RefreshCw, Trash2 } from '@lucide/svelte';
	import {
		installedApps,
		isLoading,
		viewTitle,
		appProgress,
		updateableAppsCount,
		isUpdatingAll,
		activeTasksCount,
		updateQueue,
		currentQueueAppId,
		handleUpdateAll,
		handleUpdate,
		handleUninstall,
		openDetails
	} from '$lib/stores/appStore';
</script>

<section class="flex-1 p-8 overflow-y-auto">
	<header class="mb-8 flex items-center justify-between">
		<h1 class="text-3xl font-bold tracking-tight">{$viewTitle}</h1>
		{#if $updateableAppsCount > 0}
			<button
				class="flex items-center gap-2 px-4 py-2 text-sm font-semibold bg-blue-600 text-white hover:bg-blue-700 dark:bg-blue-600 dark:hover:bg-blue-700 rounded-xl shadow-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed animate-fade-in"
				disabled={$isUpdatingAll || $activeTasksCount > 0}
				on:click={handleUpdateAll}
			>
				{#if $isUpdatingAll}
					<Loader2 class="w-4 h-4 animate-spin" />
					<span>Updating All ({$updateQueue.length} remaining)</span>
				{:else}
					<RefreshCw class="w-4 h-4" />
					<span>Update All ({$updateableAppsCount})</span>
				{/if}
			</button>
		{/if}
	</header>

	<div class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4">
		{#if $isLoading}
			<p class="col-span-full text-muted-foreground">Scanning system...</p>
		{:else if $installedApps.length === 0}
			<p class="col-span-full text-muted-foreground">No Flatpak applications installed.</p>
		{:else}
			{#each $installedApps as app}
				{@const isBusy = !!$appProgress[app.appId]}
				{@const isQueued = $isUpdatingAll && $updateQueue.includes(app.appId) && app.appId !== $currentQueueAppId}
				<article class="flex items-center justify-between bg-card border border-border p-4 rounded-xl shadow-sm">
					<div class="overflow-hidden pr-4 flex-1 cursor-pointer" on:click={() => openDetails(app.appId)}>
						<h3 class="font-semibold text-sm truncate">{app.name}</h3>
						<p class="text-xs text-muted-foreground">
							{#if isQueued}
								<span class="text-blue-500 font-medium animate-pulse">Queued for update...</span>
							{:else if isBusy}
								<span class="text-primary font-medium">{$appProgress[app.appId]?.status}... {$appProgress[app.appId]?.percentage}%</span>
							{:else}
								Version: {app.version}
							{/if}
						</p>
					</div>
					<div class="flex gap-2 shrink-0">
						{#if app.updateAvailable}
							<button
								class="p-2 bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-400 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
								title={isQueued ? 'Queued...' : isBusy ? 'Updating...' : 'Update Available'}
								disabled={isBusy || isQueued}
								on:click={() => handleUpdate(app.appId)}
							>
								{#if isBusy || isQueued}
									<Loader2 class="w-4 h-4 animate-spin" />
								{:else}
									<RefreshCw class="w-4 h-4" />
								{/if}
							</button>
						{/if}
						<button
							class="p-2 bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-900/30 dark:text-red-400 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
							title={isBusy ? 'Removing...' : 'Uninstall'}
							disabled={isBusy || isQueued}
							on:click={() => handleUninstall(app.appId)}
						>
							{#if isBusy && $appProgress[app.appId]?.status === 'removing'}
								<Loader2 class="w-4 h-4 animate-spin" />
							{:else}
								<Trash2 class="w-4 h-4" />
							{/if}
						</button>
					</div>
				</article>
			{/each}
		{/if}
	</div>
</section>
