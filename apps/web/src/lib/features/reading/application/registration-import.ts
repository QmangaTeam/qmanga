import type {ReadingGateway,ProgressStorage} from './progress';
export interface RegistrationPermit {allowed(account:string):Promise<boolean>;complete(account:string):Promise<void>}
export class RegistrationImport {
 busy=false;failed=false;
 private generation=0;private closed=false;private handledAccount='';
 constructor(private reading:ReadingGateway,private storage:ProgressStorage,private permit:RegistrationPermit,private changed:()=>void,private imported:()=>void=()=>{}){}
 clear(){this.generation++;this.handledAccount='';this.busy=false;this.failed=false;this.changed();}
 async run(){
  if(this.closed||this.busy)return;
  const generation=this.generation,valid=()=>!this.closed&&generation===this.generation;
  this.busy=true;this.failed=false;this.changed();let eligible=false,touched=false;
  try{
   const account=await this.reading.account();if(!account||!valid()||account===this.handledAccount)return;
   const allowed=await this.permit.allowed(account);if(!valid())return;if(!allowed){this.handledAccount=account;return;}
   eligible=true;const guests=await this.storage.list('guest');if(!valid())return;
   for(const guest of guests){
    try{
     const selection=await this.reading.selection(account,guest.titleId);if(!valid())return;
     // Explicit account choices take priority, including removal from the library.
     if(selection?.removed)continue;
     const cloud=await this.reading.progress(account,guest.titleId);if(!valid())return;
     if(!cloud.version){
      await this.reading.advance(account,{...this.storage.identity(),titleId:guest.titleId,releaseId:guest.progress.releaseId,revisionId:guest.progress.revisionId,pageId:guest.progress.pageId,fraction:guest.progress.fraction,baseVersion:0,completed:false});if(!valid())return;touched=true;
     }
     // A concurrent progress update is kept by the server; it is never overwritten on retry.
     if(!selection){await this.reading.save(account,{version:0,titleId:guest.titleId,status:'reading',favorite:false,preferredTeamId:'',notify:true});if(!valid())return;touched=true;}
    }catch{if(!valid())return;this.failed=true;}
   }
   if(!this.failed&&valid()){await this.permit.complete(account);if(!valid())return;this.handledAccount=account;return guests.length>0;}
  }catch{if(valid()&&eligible)this.failed=true;}
  finally{if(valid()){this.busy=false;this.changed();if(touched)this.imported();}}
 }
 close(){this.closed=true;this.clear();}
}
