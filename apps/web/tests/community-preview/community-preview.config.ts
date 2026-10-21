import {svelte,vitePreprocess} from '@sveltejs/vite-plugin-svelte';
import {defineConfig} from 'vite';
// This standalone fixture must also run before SvelteKit generates its tsconfig.
const tsconfigRaw=JSON.stringify({compilerOptions:{target:'ES2022',useDefineForClassFields:true}});
const path=(value:string)=>decodeURIComponent(new URL(value,import.meta.url).pathname);
export default defineConfig({
 esbuild:{tsconfigRaw},
 optimizeDeps:{esbuildOptions:{tsconfigRaw}},
 cacheDir:path('../../../../tmp/community-preview-vite'),
 root:path('./'),plugins:[svelte({configFile:false,preprocess:vitePreprocess()})],
 resolve:{alias:{$lib:path('../../src/lib'),'$app/state':path('./app-stubs.ts'),'$app/navigation':path('./app-stubs.ts')}},
 server:{host:'127.0.0.1',port:3002,strictPort:true,fs:{allow:[path('../../')]}}
});
