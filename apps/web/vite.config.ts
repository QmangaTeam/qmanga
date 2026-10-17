import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	build: { target: 'chrome109' },
	plugins: [sveltekit()],
	server: {
		port: 3000,
		host: true
	}
});
