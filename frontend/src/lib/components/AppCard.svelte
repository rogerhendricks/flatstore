<script lang="ts">
	import { Loader2, Check } from '@lucide/svelte';
	import type { AppSummary } from '$lib/types';
	import {
		appProgress,
		installedAppIds,
		updateableAppIds,
		isUpdatingAll,
		updateQueue,
		currentQueueAppId,
		handleImageError,
		openDetails,
		handleInstall,
		handleUpdate
	} from '$lib/stores/appStore';

	export let app: AppSummary;

	$: isBusy = !!$appProgress[app.flatpakAppId];
	$: isInstd = $installedAppIds.has(app.flatpakAppId);
	$: hasUpdate = $updateableAppIds.has(app.flatpakAppId);
	$: isQueued =
		$isUpdatingAll &&
		$updateQueue.includes(app.flatpakAppId) &&
		app.flatpakAppId !== $currentQueueAppId;
</script>

<article class="flex flex-col h-52 bg-card border border-border p-5 rounded-2xl shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all">
	<div class="flex items-center gap-4 mb-3 shrink-0 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>
		<img
			class="w-14 h-14 object-contain rounded-xl shrink-0"
			src={app.iconUrl}
			alt={app.name}
			on:error={handleImageError}
		/>
		<div class="min-w-0">
			<h3 class="font-semibold text-base leading-tight truncate">{app.name}</h3>
			<p class="text-xs text-muted-foreground mt-1 truncate">{app.developer || 'Flathub'}</p>
		</div>
	</div>
	<p class="text-sm text-muted-foreground line-clamp-3 flex-1 leading-snug mb-3 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>{app.summary}</p>
	<div class="flex justify-end mt-auto pt-2 shrink-0">
		{#if isQueued}
			<div class="flex items-center gap-1.5">
				<Loader2 class="w-3.5 h-3.5 animate-spin text-muted-foreground" />
				<span class="text-xs text-muted-foreground">Queued</span>
			</div>
		{:else if isBusy}
			<div class="flex items-center gap-1.5">
				<Loader2 class="w-3.5 h-3.5 animate-spin text-primary" />
				<span class="text-xs text-muted-foreground">{$appProgress[app.flatpakAppId]?.percentage}%</span>
			</div>
		{:else if hasUpdate}
			<button
				class="px-3 py-1.5 rounded-full text-xs font-semibold bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-400 transition-colors"
				on:click={() => handleUpdate(app.flatpakAppId)}
			>Update</button>
		{:else if isInstd}
			<Check class="w-4 h-4 text-green-500" />
		{:else}
			<button
				class="px-3 py-1.5 rounded-full text-xs font-semibold bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
				on:click={() => handleInstall(app.flatpakAppId)}
			>Get</button>
		{/if}
	</div>
</article>
