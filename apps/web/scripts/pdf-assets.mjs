import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';
import { readdirSync, readFileSync } from 'node:fs';

/** Self-host decoder data. Deliberately exclude QuickJS and JavaScript fallbacks. */
/** @returns {import('vite').Plugin} */
export function pdfAssets() {
 const root = dirname(createRequire(import.meta.url).resolve('pdfjs-dist/package.json'));
 const assets = new Map();
 for (const folder of ['cmaps', 'standard_fonts', 'wasm', 'iccs']) {
  for (const file of readdirSync(join(root, folder))) {
   if (!/\.(bcmap|pfb|ttf|wasm|icc)$/.test(file) || file.startsWith('quickjs')) continue;
   assets.set(`_pdf-assets/${folder}/${file}`, readFileSync(join(root, folder, file)));
  }
 }
 assets.set('_pdf-assets/LICENSE',readFileSync(join(root,'LICENSE')));
 for(const folder of ['wasm','standard_fonts','iccs']) for(const file of readdirSync(join(root,folder))) if(file.startsWith('LICENSE')) assets.set(`_pdf-assets/${folder}/${file}`,readFileSync(join(root,folder,file)));
 return {
  name:'qmanga-pdf-assets',
  configureServer(server) {
   server.middlewares.use((request, response, next) => {
    const key=(request.url??'').split('?')[0].replace(/^\//,'');
    const data=assets.get(key);if(!data)return next();
    response.setHeader('Content-Type',key.endsWith('.wasm')?'application/wasm':'application/octet-stream');
    response.setHeader('Cache-Control','public, max-age=3600');response.end(data);
   });
  },
  generateBundle() { for(const [fileName, source] of assets) this.emitFile({type:'asset',fileName,source}); }
 };
}
