import { MediaFailure, type Fragment, type ImageLoader, type ImageReference, type ImageTask, type Timer } from './reader';
export interface ImageState { fragment: Fragment; status: 'queued' | 'loading' | 'ready' | 'error'; image?: ImageReference; attempts: number; message: string; task?: ImageTask; stopTimer?: () => void }
export class ImageQueue {
 readonly entries = new Map<string,ImageState>();
 private order: string[] = [];
 private active = 0;
 private paused = false;
 private closed = false;
 private waiters = new Map<string,{resolve:(image:ImageReference)=>void;reject:(error:Error)=>void}[]>();
 constructor(private loader: ImageLoader, private timer: Timer, private changed: () => void, private concurrency = 2, private budget = 40*1024*1024) {}
 want(fragments: Fragment[]) {
  let memory = 0; const chosen: Fragment[] = [];
  for (const fragment of fragments) { if (chosen.some(item => item.id === fragment.id)) continue; const pixels = fragment.width*fragment.height*4; if (memory+pixels <= this.budget) { chosen.push(fragment); memory += pixels; } }
  const ids = new Set(chosen.map(item => item.id));
  for (const [id,entry] of this.entries) if (!ids.has(id)) { this.discard(entry); this.entries.delete(id); this.reject(id,new Error('Жүктеу тоқтатылды')); }
  this.order = chosen.map(fragment => fragment.id);
	for (const fragment of chosen) if (!this.entries.has(fragment.id)) {
	 const image=this.loader.cached?.(fragment);
	 this.entries.set(fragment.id,{fragment,status:image?'ready':'queued',image,attempts:0,message:''});
	}
  this.pump(); this.changed();
 }
 ready(id: string): Promise<ImageReference> { const entry = this.entries.get(id); if (entry?.image) return Promise.resolve(entry.image); if (entry?.status === 'error') return Promise.reject(new MediaFailure(entry.message,false,0,entry.message)); return new Promise((resolve,reject) => { const list=this.waiters.get(id)||[];list.push({resolve,reject});this.waiters.set(id,list); }); }
 retry(id: string) { const entry=this.entries.get(id);if(!entry)return;entry.attempts=0;entry.status='queued';entry.message='';this.pump(); }
 pause(value: boolean) { this.paused=value;if(value){for(const entry of this.entries.values()){if(entry.stopTimer){entry.stopTimer();entry.stopTimer=undefined;entry.status='queued';}}}else this.pump(); }
 private discard(entry: ImageState) { entry.stopTimer?.();if(entry.task){entry.task.cancel();entry.task=undefined;this.active--;}if(entry.image)this.loader.release(entry.image); }
 close() { this.closed=true;for(const entry of this.entries.values())this.discard(entry);this.entries.clear();for(const id of this.waiters.keys())this.reject(id,new Error('Оқу жабылды')); }
 private reject(id: string,error: Error) { for(const waiter of this.waiters.get(id)||[])waiter.reject(error);this.waiters.delete(id); }
 private pump() {
  if(this.closed||this.paused)return;
  for(const id of this.order){if(this.active>=this.concurrency)break;const entry=this.entries.get(id);if(!entry||entry.status!=='queued'||entry.stopTimer)continue;this.start(entry);}
 }
 private start(entry: ImageState) {
  entry.status='loading';this.active++;const task=this.loader.load(entry.fragment);entry.task=task;
  task.done.then(image=>{if(this.entries.get(entry.fragment.id)!==entry||entry.task!==task){this.loader.release(image);return;}entry.image=image;entry.status='ready';for(const waiter of this.waiters.get(entry.fragment.id)||[])waiter.resolve(image);this.waiters.delete(entry.fragment.id);},error=>{
   if(this.entries.get(entry.fragment.id)!==entry||entry.task!==task)return;
   const terminal=error instanceof MediaFailure&&error.terminal;
   if(!terminal&&entry.attempts<3){const delay=Math.max([1000,3000,8000][entry.attempts++],error instanceof MediaFailure?error.retryAfter:0);entry.status='queued';entry.stopTimer=this.timer.later(()=>{entry.stopTimer=undefined;this.pump();},delay);}else{entry.status='error';entry.message=error instanceof MediaFailure?error.code:'media_unavailable';this.reject(entry.fragment.id,new MediaFailure(entry.message,terminal,0,entry.message));}
  }).finally(()=>{if(entry.task===task){entry.task=undefined;this.active--;}this.changed();this.pump();});
 }
}
