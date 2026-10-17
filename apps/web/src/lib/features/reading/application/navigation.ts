import type {ReaderPreferences} from './reader';

export function navigationDirection(mode:ReaderPreferences['mode'],preference:ReaderPreferences['direction']):'ltr'|'rtl' {
 return mode==='pages'&&preference==='rtl'?'rtl':'ltr';
}

export function pageStep(side:'left'|'right',direction:'ltr'|'rtl'):number {
 return (side==='right'?1:-1)*(direction==='rtl'?-1:1);
}

export interface PageRequest {index:number;sequence:number}

export class PageNavigation {
 index=0;
 private sequence=0;
 private pending=false;
 constructor(private readonly count:number){}
 select(index:number):PageRequest|null {
  if(!Number.isInteger(index)||index<0||index>=this.count)return null;
  this.index=index;this.pending=true;
  return {index,sequence:++this.sequence};
 }
 move(delta:number):PageRequest|null {return this.select(this.index+delta);}
 current(request:PageRequest):boolean {return request.sequence===this.sequence;}
 finish(request:PageRequest):void {if(this.current(request))this.pending=false;}
 observe(index:number):void {if(!this.pending&&Number.isInteger(index)&&index>=0&&index<this.count)this.index=index;}
}
