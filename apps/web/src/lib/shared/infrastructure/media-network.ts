type Waiting = {priority:number;signal:AbortSignal;start:()=>void;reject:(reason:Error)=>void};
class MediaNetwork {
 private active=0;private queue:Waiting[]=[];
 async run<Value>(signal:AbortSignal,priority:number,operation:()=>Promise<Value>):Promise<Value>{
  if(signal.aborted)throw new Error('Жүктеу тоқтатылды');
  await new Promise<void>((resolve,reject)=>{const waiting:Waiting={priority,signal,start:resolve,reject};const abort=()=>{const index=this.queue.indexOf(waiting);if(index>=0){this.queue.splice(index,1);reject(new Error('Жүктеу тоқтатылды'));}};signal.addEventListener('abort',abort,{once:true});waiting.start=()=>{signal.removeEventListener('abort',abort);resolve();};this.queue.push(waiting);this.pump();});
  try{if(signal.aborted)throw new Error('Жүктеу тоқтатылды');return await operation();}finally{this.active--;this.pump();}
 }
 private pump(){this.queue.sort((left,right)=>left.priority-right.priority);while(this.active<2&&this.queue.length){const next=this.queue.shift()!;this.active++;next.start();}}
}
export const mediaNetwork=new MediaNetwork();
