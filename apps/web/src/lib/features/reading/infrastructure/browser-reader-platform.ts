import type { ReaderPlatform, ReaderPreferences, Viewport } from '../application/reader';
import type { Position } from '../domain/policies';
const key='qmanga:reader:preferences:v1';
export class BrowserReaderPlatform implements ReaderPlatform {
 preferences(kind: string): ReaderPreferences {
  let saved: Partial<ReaderPreferences>={};try{saved=JSON.parse(localStorage.getItem(key)||'{}');}catch{}
  return {quality:['auto','eco','standard','high'].includes(saved.quality||'')?saved.quality!:'auto',mode:saved.mode==='pages'||saved.mode==='strip'?saved.mode:kind==='manga'?'pages':'strip',theme:['dark','light','system'].includes(saved.theme||'')?saved.theme!:'dark',fit:saved.fit==='height'?'height':'width',maxWidth:Math.max(320,Math.min(1600,Number(saved.maxWidth)||1024)),brightness:Math.max(40,Math.min(100,Number(saved.brightness)||100)),numbers:saved.numbers!==false,direction:saved.direction==='ltr'||saved.direction==='rtl'?saved.direction:'auto'};
 }
 savePreferences(value: ReaderPreferences) {try{localStorage.setItem(key,JSON.stringify(value));}catch{}}
 connection() {const device=navigator as Navigator & {connection?:{saveData?:boolean;effectiveType?:string};deviceMemory?:number};return {saveData:!!device.connection?.saveData,slowConnection:['slow-2g','2g','3g'].includes(device.connection?.effectiveType||''),lowMemory:(device.deviceMemory||4)<=2};}
 observe(stage: unknown,callback: (viewport:Viewport)=>void) {
  if(!(stage instanceof HTMLElement))return ()=>{};
  let frame=0;const update=()=>{frame=0;const rect=stage.getBoundingClientRect();callback({width:rect.width,height:window.innerHeight-104,top:Math.max(0,52-rect.top)});};
  const schedule=()=>{if(!frame)frame=requestAnimationFrame(update);};const resize=new ResizeObserver(schedule);resize.observe(stage);window.addEventListener('scroll',schedule,{passive:true});window.addEventListener('resize',schedule);update();return ()=>{resize.disconnect();window.removeEventListener('scroll',schedule);window.removeEventListener('resize',schedule);cancelAnimationFrame(frame);};
 }
 scroll(stage: unknown,top:number) {if(stage instanceof HTMLElement)window.scrollTo({top:window.scrollY+stage.getBoundingClientRect().top+top-52,behavior:'instant' as ScrollBehavior});}
 panel(dialog: unknown,close:()=>void) {
  if(!(dialog instanceof HTMLDialogElement))return ()=>{};
  const marker=crypto.randomUUID();history.pushState({...history.state,qmangaPanel:marker},'');dialog.showModal();
  let finished=false;const cleanup=()=>{window.removeEventListener('popstate',pop);dialog.removeEventListener('cancel',cancel);};
  const pop=()=>{if(finished)return;finished=true;cleanup();dialog.close();close();};const cancel=(event:Event)=>{event.preventDefault();history.back();};window.addEventListener('popstate',pop);dialog.addEventListener('cancel',cancel);
  return ()=>{if(finished)return;finished=true;cleanup();dialog.close();if(history.state?.qmangaPanel===marker)history.back();};
 }
 async fullscreen(stage: unknown) {if(document.fullscreenElement)await document.exitFullscreen();else if(stage instanceof HTMLElement)await stage.requestFullscreen();}
 visibility(callback:(hidden:boolean)=>void) {const update=()=>callback(document.hidden);document.addEventListener('visibilitychange',update);return ()=>document.removeEventListener('visibilitychange',update);}
 loadPosition(releaseId:string):Position|null {try{const value=JSON.parse(localStorage.getItem('qmanga:guest-position:'+releaseId)||'null');return value&&typeof value.pageId==='string'&&Number.isFinite(value.fraction)?{pageId:value.pageId,fraction:Math.max(0,Math.min(1,value.fraction))}:null;}catch{return null;}}
 savePosition(releaseId:string,position:Position) {try{localStorage.setItem('qmanga:guest-position:'+releaseId,JSON.stringify(position));}catch{}}
 keyboard(callback:(action:'next'|'previous'|'menu')=>void) {const handle=(event:KeyboardEvent)=>{if(event.target instanceof HTMLElement&&(event.target.closest('input,textarea,select,button,a,[contenteditable],dialog')))return;const action=event.key==='ArrowRight'?'next':event.key==='ArrowLeft'?'previous':event.key==='m'?'menu':null;if(action){event.preventDefault();callback(action);}};window.addEventListener('keydown',handle);return ()=>window.removeEventListener('keydown',handle);}
}
