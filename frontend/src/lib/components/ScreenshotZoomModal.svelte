<script lang="ts">
	import { fade } from 'svelte/transition';
	import { zoomedScreenshot } from '$lib/stores/appStore';
</script>

{#if $zoomedScreenshot}
	<div
		class="fixed inset-0 z-50 bg-black/90 flex items-center justify-center p-8 cursor-zoom-out"
		on:click={() => zoomedScreenshot.set(null)}
		transition:fade={{ duration: 150 }}
	>
		<img
			class="max-w-full max-h-full rounded-xl object-contain shadow-2xl border border-white/10"
			src={$zoomedScreenshot}
			alt="Zoomed Screenshot"
		/>
	</div>
{/if}
