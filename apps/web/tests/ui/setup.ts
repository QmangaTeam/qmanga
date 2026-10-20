import { plugin } from 'bun';
import { existsSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { compile } from 'svelte/compiler';
import { mock } from 'bun:test';

mock.module('$app/navigation', () => ({ goto: async () => {}, invalidateAll: async () => {}, beforeNavigate: () => {}, afterNavigate: () => {}, replaceState: () => {} }));
mock.module('$app/state', () => ({ page: { url: new URL('http://localhost/') } }));

plugin({
 name: 'svelte-ui-contracts',
 setup(build) {
  build.onResolve({ filter: /^\$lib\// }, ({ path }) => {
   const source = resolve('src/lib', path.slice(5));
   const target = [source, source + '.ts', source + '.svelte', source + '/index.ts'].find(filename => existsSync(filename));
   if (!target) throw new Error('Unresolved UI dependency: ' + path);
   return { path: target };
  });
  build.onResolve({ filter: /^\$app\// }, () => ({ path: resolve('tests/ui/app-stubs.ts') }));
  build.onLoad({ filter: /\.svelte$/ }, ({ path }) => ({
   contents: compile(readFileSync(path, 'utf8'), { filename: path, generate: 'server' }).js.code,
   loader: 'js'
  }));
 }
});
