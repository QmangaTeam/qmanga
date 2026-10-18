import type {Notification,NotificationsGateway} from './notifications';
function code(error:unknown){return error&&typeof error==='object'&&'code' in error?String(error.code):'notification_failed';}
export class NotificationFeed {
 account:string|null=null;items:Notification[]=[];busy=false;more=false;error='';private generation=0;private closed=false;
 constructor(private gateway:NotificationsGateway,private changed:()=>void=()=>{}){}
 reset(){this.generation++;this.account=null;this.items=[];this.busy=false;this.more=false;this.error='';this.changed();}
 private accountChanged(){this.reset();this.error='account_changed';this.changed();}
 async load(after=''){
  if(this.busy||this.closed)return;const generation=this.generation;this.busy=true;this.error='';this.changed();
  try{
   const account=await this.gateway.account();if(this.closed||generation!==this.generation)return;
   if(account!==this.account){this.items=[];this.more=false;this.account=account;after='';this.changed();}
   if(!account)return;
   const result=await this.gateway.list(account,after);
   if(this.closed||generation!==this.generation)return;
   const current=await this.gateway.account();if(this.closed||generation!==this.generation)return;
   if(current!==account){this.accountChanged();return;}
   this.items=after?[...this.items,...result.items]:result.items;this.more=result.hasMore;
  }catch(error){if(!this.closed&&generation===this.generation)this.error=code(error);}
  finally{if(!this.closed&&generation===this.generation){this.busy=false;this.changed();}}
 }
 async read(item:Notification){
  const account=this.account,generation=this.generation;if(!account||this.closed||!this.items.some(value=>value.id===item.id))return;
  try{
   const current=await this.gateway.account();if(this.closed||generation!==this.generation)return;
   if(current!==account){this.accountChanged();return;}
   await this.gateway.read(account,item.id);
   const latest=await this.gateway.account();if(this.closed||generation!==this.generation)return;
   if(latest!==account){this.accountChanged();return;}
   this.items=this.items.map(value=>value.id===item.id?{...value,read:true}:value);this.changed();
  }catch(error){if(!this.closed&&generation===this.generation){this.error=code(error);this.changed();}}
 }
 close(){this.closed=true;this.reset();}
}
