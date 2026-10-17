export interface MediaCache {
 match(key:string|Request):Promise<Response|undefined>;
 put(key:string|Request,response:Response):Promise<void>;
 delete(key:string|Request):Promise<boolean>;
 keys():Promise<readonly Request[]>;
}
export class RuntimeMediaCache {
 private chain:Promise<unknown>=Promise.resolve();
 constructor(private cache:MediaCache,private now:()=>number,private budget=50*1024*1024){}
 private serialized<Value>(action:()=>Promise<Value>):Promise<Value>{const result=this.chain.then(action);this.chain=result.catch(()=>{});return result;}
 store(key:string,response:Response):Promise<void>{return this.serialized(async()=>{
  const policy=response.headers.get('cache-control')||'';
  const mime=response.headers.get('content-type')?.split(';')[0];
  const size=Number(response.headers.get('content-length'));
  if(response.status!==200||!/(?:^|,)\s*public(?:\s*,|\s*$)/i.test(policy)||/private|no-store/i.test(policy)||!mime||!['image/webp','image/avif','image/jpeg','image/png'].includes(mime)||!Number.isSafeInteger(size)||size<=0||size>this.budget)return;
  const body=await response.arrayBuffer();if(body.byteLength!==size)return;
  const rows:Array<{key:Request;size:number;used:number}>=[];let total=0;
  for(const request of await this.cache.keys()){
   if(request.url===key)continue;
   const entry=await this.cache.match(request),bytes=Number(entry?.headers.get('content-length')),used=Number(entry?.headers.get('x-qmanga-accessed'));
   if(!Number.isSafeInteger(bytes)||bytes<=0||!Number.isFinite(used)){await this.cache.delete(request);continue;}
   total+=bytes;rows.push({key:request,size:bytes,used});
  }
  rows.sort((left,right)=>left.used-right.used);
  for(const entry of rows){if(total+size<=this.budget)break;await this.cache.delete(entry.key);total-=entry.size;}
  const headers=new Headers(response.headers);headers.set('x-qmanga-accessed',String(this.now()));
  await this.cache.put(key,new Response(body,{status:200,headers}));
 });}
 read(key:string):Promise<Response|undefined>{return this.serialized(async()=>{
  const entry=await this.cache.match(key);if(!entry)return;
  const headers=new Headers(entry.headers);headers.set('x-qmanga-accessed',String(this.now()));
  await this.cache.put(key,new Response(entry.clone().body,{status:200,headers})).catch(()=>{});return entry;
 });}
 remove(key:string):Promise<void>{return this.serialized(async()=>{await this.cache.delete(key);});}
}
