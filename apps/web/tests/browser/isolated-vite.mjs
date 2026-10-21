import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

// These codec servers run on a cold host after Docker checks. They must not
// depend on generated SvelteKit config or write into Docker-owned node_modules.
export async function isolatedViteOptions() {
 const tsconfigRaw = JSON.stringify({ compilerOptions: { target: 'ES2022', useDefineForClassFields: true } });
 return {
  appType: 'custom',
  cacheDir: await mkdtemp(join(tmpdir(), 'qmanga-codec-vite-')),
  esbuild: { tsconfigRaw },
  optimizeDeps: { entries: [], esbuildOptions: { tsconfigRaw } },
 };
}

export async function cleanViteCache(server) {
 if (server?.config.cacheDir) await rm(server.config.cacheDir, { recursive: true, force: true });
}
