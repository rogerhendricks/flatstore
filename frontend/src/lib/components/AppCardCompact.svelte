<script lang="ts">
	import { Loader2, Check } from '@lucide/svelte';
	import type { AppSummary } from '$lib/types';
	import {
		appProgress,
		installedAppIds,
		updateableAppIds,
		handleImageError,
		openDetails,
		handleInstall,
		handleUpdate
	} from '$lib/stores/appStore';

	export let app: AppSummary;
	export let bordered: boolean = false;

	$: isBusy = !!$appProgress[app.flatpakAppId];
	$: isInstd = $installedAppIds.has(app.flatpakAppId);
	$: hasUpdate = $updateableAppIds.has(app.flatpakAppId);
</script>

<div class="flex flex-col justify-between px-3 py-2.5 h-24 hover:bg-muted/50 transition-colors {bordered ? 'border-t border-border' : ''}">
	<div class="flex items-start gap-2.5 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>
		<img
			class="w-10 h-10 rounded-xl object-contain shrink-0"
			src={app.iconUrl}
			alt={app.name}
			on:error={handleImageError}
		/>
		<div class="flex-1 overflow-hidden">
			<p class="text-xs font-semibold leading-tight line-clamp-1 text-zinc-800 dark:text-zinc-200">{app.name}</p>
			<p class="text-[11px] leading-snug mt-0.5 line-clamp-2 text-zinc-500 dark:text-zinc-400">{app.summary}</p>
		</div>
	</div>
	<div class="flex justify-end">
		{#if isBusy}
			<div class="flex items-center gap-1">
				<Loader2 class="w-3 h-3 animate-spin text-primary" />
				<span class="text-[10px] text-muted-foreground">{$appProgress[app.flatpakAppId]?.percentage}%</span>
			</div>
		{:else if hasUpdate}
			<button
				class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-400 transition-colors"
				on:click={() => handleUpdate(app.flatpakAppId)}
			>Update</button>
		{:else if isInstd}
			<Check class="w-4 h-4 text-green-500" />
		{:else}
			<button
				class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
				on:click={() => handleInstall(app.flatpakAppId)}
			>Get</button>
		{/if}
	</div>
</div>
