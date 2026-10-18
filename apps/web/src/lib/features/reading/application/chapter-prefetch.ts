import type {ImageLoader,Manifest,Tier} from './reader';

// Preload only a bounded amount; foreground reader requests always have priority.
export async function prefetchChapter(manifest:Manifest,tier:Tier,loader:ImageLoader,signal:AbortSignal,budget=24*1024*1024):Promise<void>{
 let used=0;
 for(const page of manifest.pages){
  for(const fragment of page.variants[tier]){
   if(signal.aborted||used+fragment.bytes>budget)return;
   used+=fragment.bytes;
   const task=loader.load(fragment,5);
   const abort=()=>task.cancel();signal.addEventListener('abort',abort,{once:true});
   try{const image=await task.done;loader.release(image);}catch{return;}finally{signal.removeEventListener('abort',abort);}
  }
 }
}
