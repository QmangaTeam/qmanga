import {svelte,vitePreprocess} from '@sveltejs/vite-plugin-svelte';
import {defineConfig} from 'vite';
const path=(value:string)=>decodeURIComponent(new URL(value,import.meta.url).pathname);
export default defineConfig({
 cacheDir:path('../../../../tmp/community-preview-vite'),
 root:path('./'),plugins:[svelte({configFile:false,preprocess:vitePreprocess()})],
 resolve:{alias:{$lib:path('../../src/lib'),'$app/state':path('./app-stubs.ts'),'$app/navigation':path('./app-stubs.ts')}},
 server:{host:'127.0.0.1',port:3002,strictPort:true,fs:{allow:[path('../../')]}}
});
