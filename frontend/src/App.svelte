<script lang="ts">
	import './app.css';
	import { onMount, onDestroy } from 'svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import DiscoverView from '$lib/components/DiscoverView.svelte';
	import PopularView from '$lib/components/PopularView.svelte';
	import InstalledView from '$lib/components/InstalledView.svelte';
	import AppDetailsPage from '$lib/components/AppDetailsPage.svelte';
	import ScreenshotZoomModal from '$lib/components/ScreenshotZoomModal.svelte';
	import { activeCategory, selectedAppIdForPage, initApp, destroyApp } from '$lib/stores/appStore';

	onMount(() => {
		initApp();
	});

	onDestroy(() => {
		destroyApp();
	});
</script>

<main class="flex h-screen w-screen overflow-hidden bg-background text-foreground select-none">
	<Sidebar />

	{#if $selectedAppIdForPage}
		<AppDetailsPage />
	{:else}
		{#if $activeCategory !== 'Installed' && $activeCategory !== 'Popular'}
			<DiscoverView />
		{/if}
		{#if $activeCategory === 'Popular'}
			<PopularView />
		{/if}
		{#if $activeCategory === 'Installed'}
			<InstalledView />
		{/if}
	{/if}

	<ScreenshotZoomModal />
</main>
