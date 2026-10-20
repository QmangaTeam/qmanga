import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { pdfAssets } from './scripts/pdf-assets.mjs';

export default defineConfig({
	build: { target: 'chrome109' },
	plugins: [pdfAssets(), sveltekit()],
	server: {
		port: 3000,
		host: true
	}
});
