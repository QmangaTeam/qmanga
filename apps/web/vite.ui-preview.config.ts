import { svelte, vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

const localPath = (relative: string) => decodeURIComponent(new URL(relative, import.meta.url).pathname).replace(/^\/([A-Za-z]:\/)/, '$1');

export default defineConfig({
 root: localPath('./ui-preview'),
 plugins: [svelte({ configFile: false, preprocess: vitePreprocess() })],
 resolve: { alias: { $lib: localPath('./src/lib') } },
 build: { target: 'chrome109' },
 server: { host: '127.0.0.1', port: 3001, strictPort: true, fs: { allow: [localPath('./')] } }
});
