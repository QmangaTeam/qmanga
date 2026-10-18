import type {ReadingGateway} from './progress';
export interface SuggestionSession {libraryPromptDismissed?(account:string,title:string):boolean;dismissLibraryPrompt?(account:string,title:string):void}
export class LibrarySuggestion {
 visible=false;busy=false;error='';private checked=false;private account='';private stopped=false;
 constructor(private gateway:ReadingGateway,private session:SuggestionSession,private title:string,private changed:()=>void=()=>{}){}
 async check(account:string,accepted:boolean){
  if(this.stopped||this.checked||!accepted||account==='guest'||!this.gateway.suggestion||this.session.libraryPromptDismissed?.(account,this.title))return;
  this.checked=true;
  try{const eligible=await this.gateway.suggestion(account,this.title);if(!this.stopped&&eligible){this.account=account;this.visible=true;this.changed();}}catch{this.checked=false;}
 }
 dismiss(){this.visible=false;this.session.dismissLibraryPrompt?.(this.account,this.title);this.changed();}
 async add(){
  if(this.busy||!this.account||this.stopped)return;this.busy=true;this.error='';this.changed();
  try{
   if(await this.gateway.account()!==this.account)throw {code:'account_changed'};
   const entry=await this.gateway.selection(this.account,this.title);
   if(!entry||entry.removed)await this.gateway.save(this.account,{version:entry?.version||0,titleId:this.title,status:'reading',favorite:false,preferredTeamId:'',notify:true});
   if(!this.stopped)this.dismiss();
  }catch(failure){this.error=typeof failure==='object'&&failure!==null&&'code' in failure?String(failure.code):'library_failed';}
  finally{this.busy=false;this.changed();}
 }
 close(){this.stopped=true;this.visible=false;}
}
