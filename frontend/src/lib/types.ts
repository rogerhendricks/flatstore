export interface AppDetails {
	flatpakAppId: string;
	name: string;
	summary: string;
	description: string;
	homepageUrl: string;
	bugtrackerUrl: string;
	helpUrl: string;
	vcsBrowserUrl: string;
	iconUrl: string;
	version: string;
	developer: string;
	verified: boolean;
	screenshots: string[];
	releaseDate: string;
	ageRating: string;
	license: string;
	releases: Release[];
}

export interface Release {
	version: string;
	date: string;
	description: string;
}

export interface AppSummary {
	flatpakAppId: string;
	name: string;
	summary: string;
	iconUrl: string;
	version: string;
	developer: string;
}

export type ProgressStatus =
	| 'starting'
	| 'downloading'
	| 'installing'
	| 'completed'
	| 'removing'
	| 'error';

export interface InstalledApp {
	appId: string;
	name: string;
	version: string;
	updateAvailable: boolean;
}

export interface AppProgress {
	status: ProgressStatus;
	percentage: number;
	name?: string;
}

export interface ProgressPayload {
	appId: string;
	status: ProgressStatus;
	percentage: number;
}

export interface Category {
	id: string;
	label: string;
	icon: any; // Svelte component type
}

export type Theme = 'light' | 'dark' | 'system';

export type PopularTab = 'discover' | 'games' | 'create';

export interface FeaturedPromo {
	badge: string;
	title: string;
	subtitle: string;
	gradient: { from: string; to: string };
}

export interface HeroGradient {
	from: string;
	via: string;
	to: string;
}

export type CatalogStatus = 'syncing' | 'ready' | 'error';
