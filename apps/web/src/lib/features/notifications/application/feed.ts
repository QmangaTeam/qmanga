import type {Notification,NotificationsGateway} from './notifications';
function code(error:unknown){return error&&typeof error==='object'&&'code' in error?String(error.code):'notification_failed';}
export class NotificationFeed {
 kind:''|'chapters'|'replies'='';account:string|null=null;items:Notification[]=[];busy=false;more=false;error='';unreadOnly=false;cursor='';private generation=0;private closed=false;
 constructor(private gateway:NotificationsGateway,private changed:()=>void=()=>{}){}
 reset(){this.generation++;this.account=null;this.items=[];this.busy=false;this.more=false;this.cursor='';this.error='';this.changed();}
 private accountChanged(){this.reset();this.error='account_changed';this.changed();}
 async setUnread(value:boolean){await this.setFilter(value,this.kind);}
 async setFilter(unread:boolean,kind:''|'chapters'|'replies'){if(unread===this.unreadOnly&&kind===this.kind)return;this.reset();this.unreadOnly=unread;this.kind=kind;await this.load();}
 async load(after=''){
  if(this.busy||this.closed)return;const generation=this.generation;this.busy=true;this.error='';this.changed();
  try{
   const account=await this.gateway.account();if(this.closed||generation!==this.generation)return;
   if(account!==this.account){this.items=[];this.more=false;this.cursor='';this.account=account;after='';this.changed();}
   if(!account)return;
   const result=await this.gateway.list(account,after,this.unreadOnly,this.kind);
   if(this.closed||generation!==this.generation)return;
   const current=await this.gateway.account();if(this.closed||generation!==this.generation)return;
   if(current!==account){this.accountChanged();return;}
   this.items=after?[...this.items,...result.items]:result.items;this.more=result.hasMore;this.cursor=result.items.at(-1)?.id||after;
  }catch(error){if(!this.closed&&generation===this.generation)this.error=code(error);}
  finally{if(!this.closed&&generation===this.generation){this.busy=false;this.changed();}}
 }
 async read(item:Notification){
  const account=this.account,generation=this.generation;if(!account||this.closed||!this.items.some(value=>value.id===item.id))return;
  try{
   const current=await this.gateway.account();if(!this.closed&&generation!==this.generation)return;
   if(current!==account){if(!this.closed)this.accountChanged();return;}
   // A link may unmount the view while identity is checked; finish the user's
   // already-started read, with the server's expected-account guard intact.
   await this.gateway.read(account,item.id);
   if(this.closed||generation!==this.generation)return;
   const latest=await this.gateway.account();if(this.closed||generation!==this.generation)return;
   if(latest!==account){this.accountChanged();return;}
   this.items=this.unreadOnly?this.items.filter(value=>value.id!==item.id):this.items.map(value=>value.id===item.id?{...value,read:true}:value);this.changed();
  }catch(error){if(!this.closed&&generation===this.generation){this.error=code(error);this.changed();}}
 }
 async readAll(){
  const account=this.account,generation=this.generation;if(!account||this.closed||this.busy||!this.gateway.readAll)return;
  this.busy=true;this.error='';this.changed();
  try{if(await this.gateway.account()!==account){if(generation===this.generation)this.accountChanged();return;}
   if(this.closed||generation!==this.generation)return;
   await this.gateway.readAll(account);
   if(this.closed||generation!==this.generation)return;
   const latest=await this.gateway.account();if(this.closed||generation!==this.generation)return;if(latest!==account){this.accountChanged();return;}
   this.items=this.unreadOnly?[]:this.items.map(item=>({...item,read:true}));if(this.unreadOnly){this.more=false;this.cursor='';}this.changed();
  }catch(error){if(!this.closed&&generation===this.generation){this.error=code(error);this.changed();}}
  finally{if(!this.closed&&generation===this.generation){this.busy=false;this.changed();}}
 }
 close(){this.closed=true;this.reset();}
}
