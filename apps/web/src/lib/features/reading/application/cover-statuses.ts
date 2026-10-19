import type {ReadingGateway} from './progress';
export class CoverStatuses {
 items:Record<string,string>={};
 private generation=0;
 private closed=false;
 constructor(private gateway:ReadingGateway,private changed:()=>void){}
 clear(){this.generation++;this.items={};this.changed();}
 async load(ids:string[]){
  if(this.closed)return;
  this.clear();const generation=this.generation;
  try{
   const account=await this.gateway.account();
   if(!account||!this.gateway.statuses||this.closed||generation!==this.generation)return;
   const unique=[...new Set(ids)],result:Record<string,string>={};
   for(let offset=0;offset<unique.length;offset+=60){
    if(this.closed||generation!==this.generation)return;
    Object.assign(result,await this.gateway.statuses(account,unique.slice(offset,offset+60)));
   }
   if(this.closed||generation!==this.generation)return;
   this.items=result;this.changed();
  }catch{/* A failed personal request must never invent or reuse another account's status. */}
 }
 close(){this.closed=true;this.clear();}
}
