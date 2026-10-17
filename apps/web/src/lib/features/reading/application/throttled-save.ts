import type {Timer} from './reader';
export class ThrottledSave<Value>{
 private pending:Value|undefined;private cancel:(()=>void)|undefined;private writing:Promise<void>=Promise.resolve();
 constructor(private timer:Timer,private write:(value:Value)=>Promise<void>){}
 offer(value:Value){this.pending=value;if(!this.cancel)this.cancel=this.timer.later(()=>{this.cancel=undefined;void this.flush();},2000);}
 flush():Promise<void>{this.cancel?.();this.cancel=undefined;const value=this.pending;this.pending=undefined;if(value!==undefined)this.writing=this.writing.catch(()=>{}).then(()=>this.write(value));return this.writing;}
}
