import {
	MonitorPlay,
	Code,
	GraduationCap,
	Gamepad2,
	Palette,
	Globe,
	Briefcase,
	FlaskConical,
	Wrench
} from '@lucide/svelte';
import type { Category, FeaturedPromo, HeroGradient } from '$lib/types';

export const categories: Category[] = [
	{ id: 'AudioVideo', label: 'Audio & Video', icon: MonitorPlay },
	{ id: 'Development', label: 'Development', icon: Code },
	{ id: 'Education', label: 'Education', icon: GraduationCap },
	{ id: 'Game', label: 'Games', icon: Gamepad2 },
	{ id: 'Graphics', label: 'Graphics', icon: Palette },
	{ id: 'Network', label: 'Internet', icon: Globe },
	{ id: 'Office', label: 'Office', icon: Briefcase },
	{ id: 'Science', label: 'Science', icon: FlaskConical },
	{ id: 'System', label: 'System', icon: Wrench }
];

// Gradient definitions for hero cards (as inline CSS to avoid Tailwind purging).
export const heroGradients: HeroGradient[] = [
	{ from: '#7c3aed', via: '#9333ea', to: '#3b82f6' }, // violet → blue
	{ from: '#f97316', via: '#ef4444', to: '#ec4899' }, // orange → pink
	{ from: '#10b981', via: '#14b8a6', to: '#06b6d4' }, // emerald → cyan
	{ from: '#3b82f6', via: '#6366f1', to: '#8b5cf6' }, // blue → purple
	{ from: '#f43f5e', via: '#f97316', to: '#f59e0b' } // rose → amber
];

// Placeholder featured promo cards — populated by external API in a future iteration.
export const featuredPromos: FeaturedPromo[] = [
	{
		badge: 'Featured',
		title: "Editor's Choice",
		subtitle: 'Hand-picked highlights from our team',
		gradient: { from: '#5b21b6', to: '#7c3aed' }
	},
	{
		badge: 'New',
		title: 'Fresh Arrivals',
		subtitle: 'Brand new apps just added to Flathub',
		gradient: { from: '#9d174d', to: '#e11d48' }
	},
	{
		badge: 'Staff Pick',
		title: 'Must-Have Tools',
		subtitle: 'Essential apps for your Linux desktop',
		gradient: { from: '#075985', to: '#0ea5e9' }
	}
];
