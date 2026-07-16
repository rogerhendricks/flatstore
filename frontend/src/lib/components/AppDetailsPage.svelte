<script lang="ts">
	import { fade } from 'svelte/transition';
	import { LoaderCircle, ChevronLeft, ExternalLink, BadgeCheck } from '@lucide/svelte';
	import { Progress } from '$lib/components/ui/progress';
	import {
		selectedAppDetails,
		isDetailsLoading,
		appProgress,
		installedAppIds,
		updateableAppIds,
		isUpdatingAll,
		updateQueue,
		currentQueueAppId,
		closeDetails,
		handleImageError,
		handleUpdate,
		handleOpenApp,
		handleUninstall,
		handleInstall,
		getProgressColorClass,
		zoomedScreenshot,
		loadByDeveloper
	} from '$lib/stores/appStore';
</script>

<section class="flex-1 overflow-y-auto p-8 bg-background" transition:fade={{ duration: 150 }}>
	{#if $isDetailsLoading}
		<div class="flex flex-col items-center justify-center h-[80vh] text-muted-foreground gap-3">
			<LoaderCircle class="w-8 h-8 animate-spin text-primary" />
			<p class="text-sm">Loading application details...</p>
		</div>
	{:else if $selectedAppDetails}
		{@const app = $selectedAppDetails}
		{@const isBusy = !!$appProgress[app.flatpakAppId]}
		{@const isInstd = $installedAppIds.has(app.flatpakAppId)}
		{@const hasUpdate = $updateableAppIds.has(app.flatpakAppId)}
		{@const isQueued = $isUpdatingAll && $updateQueue.includes(app.flatpakAppId) && app.flatpakAppId !== $currentQueueAppId}

		<!-- Back Button -->
		<button
			class="flex items-center gap-2 text-muted-foreground hover:text-foreground text-sm font-medium mb-6 transition-colors"
			on:click={closeDetails}
		>
			<ChevronLeft class="w-4 h-4" />
			<span>Back</span>
		</button>

		<!-- Hero Header Card -->
		<div class="flex items-start gap-6 mb-6">
			<img
				class="w-24 h-24 rounded-2xl object-contain bg-card border border-border p-3 shrink-0 shadow-sm"
				src={app.iconUrl}
				alt={app.name}
				on:error={handleImageError}
			/>
			<div class="min-w-0 flex-1">
				<h2 class="text-3xl font-extrabold tracking-tight text-foreground leading-tight">{app.name}</h2>
				{#if app.developer}
					<div class="flex items-center gap-1 mt-1">
						{#if app.verified}
							<BadgeCheck class="w-4 h-4 text-blue-500 shrink-0" />
						{/if}
						<button
							class="text-sm text-primary font-semibold hover:underline"
							on:click={() => loadByDeveloper(app.developer)}
						>{app.developer}</button>
					</div>
				{:else}
					<p class="text-sm text-primary font-semibold mt-1">Flathub</p>
				{/if}
				<p class="text-sm text-muted-foreground mt-2 leading-relaxed max-w-2xl">{app.summary}</p>
			</div>
		</div>

		<!-- Actions Bar -->
		<div class="flex items-center gap-3 mb-8 max-w-sm shrink-0">
			{#if isQueued}
				<button
					class="flex-1 flex items-center justify-center gap-2 py-2 px-4 rounded-xl text-sm font-semibold bg-muted text-muted-foreground cursor-not-allowed border border-border"
					disabled
				>
					<LoaderCircle class="w-4 h-4 animate-spin text-muted-foreground" />
					<span>Queued</span>
				</button>
			{:else if isBusy}
				<div class="flex-1 flex flex-col gap-1.5 bg-muted/30 p-2.5 rounded-xl border border-border font-semibold text-xs text-primary">
					<div class="flex justify-between px-1 mb-1">
						<span class="capitalize">{$appProgress[app.flatpakAppId]?.status}...</span>
						<span>{$appProgress[app.flatpakAppId]?.percentage}%</span>
					</div>
					<Progress value={$appProgress[app.flatpakAppId]?.percentage} class="h-2 {getProgressColorClass($appProgress[app.flatpakAppId]?.status)}" />
				</div>
			{:else if hasUpdate}
				<button
					class="flex-1 py-2 px-4 rounded-xl text-sm font-semibold bg-blue-600 hover:bg-blue-700 text-white shadow-sm transition-colors"
					on:click={() => handleUpdate(app.flatpakAppId)}
				>
					Update
				</button>
				<button
					class="py-2 px-4 bg-muted hover:bg-muted/80 text-muted-foreground rounded-xl transition-colors text-sm font-semibold"
					on:click={() => handleOpenApp(app.flatpakAppId)}
				>
					Open
				</button>
				<button
					class="py-2 px-4 bg-red-100 hover:bg-red-200 text-red-700 dark:bg-red-900/30 dark:hover:bg-red-900/50 dark:text-red-400 rounded-xl transition-colors text-sm font-semibold"
					on:click={() => handleUninstall(app.flatpakAppId)}
				>
					Uninstall
				</button>
			{:else if isInstd}
				<div class="flex gap-3 w-full">
					<button
						class="flex-1 py-2 px-4 rounded-xl text-sm font-semibold bg-primary hover:bg-primary/95 text-primary-foreground shadow-sm transition-colors flex items-center justify-center gap-1.5"
						on:click={() => handleOpenApp(app.flatpakAppId)}
					>
						Open
					</button>
					<button
						class="py-2 px-4 bg-red-100 hover:bg-red-200 text-red-700 dark:bg-red-900/30 dark:hover:bg-red-900/50 dark:text-red-400 rounded-xl transition-colors text-sm font-semibold"
						on:click={() => handleUninstall(app.flatpakAppId)}
					>
						Uninstall
					</button>
				</div>
			{:else}
				<button
					class="flex-1 py-2 px-4 rounded-xl text-sm font-semibold bg-primary hover:bg-primary/95 text-primary-foreground shadow-sm transition-colors"
					on:click={() => handleInstall(app.flatpakAppId)}
				>
					Get
				</button>
			{/if}
		</div>

		<!-- Metadata Section -->
		<hr class="border-border my-6" />
		<div class="grid grid-cols-4 gap-y-6 gap-x-4 text-xs max-w-4xl mb-8">
			<div>
				<p class="text-muted-foreground font-semibold uppercase tracking-wider mb-1">Developer</p>
				<p class="text-sm font-bold text-foreground truncate">{app.developer || 'Flathub'}</p>
			</div>
			<div>
				<p class="text-muted-foreground font-semibold uppercase tracking-wider mb-1">Version</p>
				<p class="text-sm font-bold text-foreground truncate">{app.version || 'Unknown'}</p>
			</div>
			{#if app.releaseDate}
				<div>
					<p class="text-muted-foreground font-semibold uppercase tracking-wider mb-1">Released On</p>
					<p class="text-sm font-bold text-foreground truncate">{app.releaseDate}</p>
				</div>
			{/if}
			<div>
				<p class="text-muted-foreground font-semibold uppercase tracking-wider mb-1">Age Rating</p>
				<p class="text-sm font-bold text-foreground truncate">{app.ageRating || 'Everyone'}</p>
			</div>
			<div>
				<p class="text-muted-foreground font-semibold uppercase tracking-wider mb-1">License</p>
				<p class="text-sm font-bold text-foreground truncate" title={app.license}>{app.license || 'Unknown'}</p>
			</div>
			{#if app.homepageUrl}
				<div class="col-span-4 mt-2">
					<p class="text-muted-foreground font-semibold uppercase tracking-wider mb-2">Project Links</p>
					<div class="flex gap-4">
						<a
							href={app.homepageUrl}
							target="_blank"
							class="flex items-center gap-1 text-primary hover:underline font-bold"
						>
							<span>Homepage</span>
							<ExternalLink class="w-3.5 h-3.5" />
						</a>
						{#if app.bugtrackerUrl}
							<a
								href={app.bugtrackerUrl}
								target="_blank"
								class="flex items-center gap-1 text-primary hover:underline font-bold"
							>
								<span>Bug Tracker</span>
								<ExternalLink class="w-3.5 h-3.5" />
							</a>
						{/if}
					</div>
				</div>
			{/if}
		</div>

		<!-- Screenshots Section -->
		{#if app.screenshots && app.screenshots.length > 0}
			<hr class="border-border my-6" />
			<h3 class="text-base font-bold text-foreground mb-4">Screenshots</h3>
			<div class="flex gap-4 overflow-x-auto pb-4 scrollbar-thin scrollbar-thumb-rounded max-w-4xl">
				{#each app.screenshots as src}
					<img
						class="h-64 rounded-xl border border-border object-cover cursor-zoom-in hover:brightness-95 transition-all shadow-sm shrink-0"
						{src}
						alt="Screenshot of {app.name}"
						on:click={() => zoomedScreenshot.set(src)}
					/>
				{/each}
			</div>
		{/if}

		<!-- Description Section -->
		<hr class="border-border my-6" />
		<h3 class="text-base font-bold text-foreground mb-4">About</h3>
		<div class="text-sm leading-relaxed text-muted-foreground space-y-4 prose dark:prose-invert max-w-3xl pb-16">
			{@html app.description}
		</div>
	{/if}
</section>
