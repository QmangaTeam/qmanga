import {test,expect} from 'bun:test';
import {BrowserReaderPlatform} from './browser-reader-platform';

test('arrow keys work after focusing reader buttons and links, without hijacking editing or shortcuts',()=>{
 const oldWindow=globalThis.window,oldElement=globalThis.HTMLElement;
 class FakeElement extends EventTarget {
  constructor(readonly tag:string){super();}
  closest(selector:string){return selector.split(',').some(value=>value.trim()===this.tag)?this:null;}
 }
 const target=new EventTarget(),actions:string[]=[];
 globalThis.window=target as unknown as Window & typeof globalThis;
 globalThis.HTMLElement=FakeElement as unknown as typeof HTMLElement;
 const press=(tag:string,key:string,options:Record<string,boolean>={})=>{
  const event=new Event('keydown',{cancelable:true});
  Object.defineProperties(event,{target:{value:new FakeElement(tag)},key:{value:key},...Object.fromEntries(Object.entries(options).map(([name,value])=>[name,{value}]))});
  target.dispatchEvent(event);return event.defaultPrevented;
 };
 try{
  const stop=new BrowserReaderPlatform().keyboard(action=>actions.push(action));
  expect(press('button','ArrowRight')).toBe(true);
  expect(press('a','ArrowLeft')).toBe(true);
  for(const tag of ['input','textarea','select','[contenteditable]','dialog'])expect(press(tag,'ArrowRight')).toBe(false);
  for(const modifier of ['ctrlKey','altKey','metaKey','shiftKey','isComposing'])expect(press('button','ArrowRight',{[modifier]:true})).toBe(false);
  expect(press('button','Enter')).toBe(false);
  expect(press('button',' ')).toBe(false);
  expect(actions).toEqual(['right','left']);
  stop();press('button','ArrowRight');expect(actions).toHaveLength(2);
 }finally{globalThis.window=oldWindow;globalThis.HTMLElement=oldElement;}
});
test('Back closes the drawer and removes its history listener exactly once',()=>{
 const oldWindow=globalThis.window,oldHistory=globalThis.history,oldDialog=globalThis.HTMLDialogElement;let listeners=0,closed=0;
 class FakeWindow extends EventTarget{addEventListener(type:string,listener:EventListenerOrEventListenerObject,options?:AddEventListenerOptions|boolean){if(type==='popstate')listeners++;super.addEventListener(type,listener,options);}removeEventListener(type:string,listener:EventListenerOrEventListenerObject,options?:EventListenerOptions|boolean){if(type==='popstate')listeners--;super.removeEventListener(type,listener,options);}}
 class FakeDialog extends EventTarget{showModal(){}close(){}}
 globalThis.window=new FakeWindow() as unknown as Window & typeof globalThis;globalThis.HTMLDialogElement=FakeDialog as unknown as typeof HTMLDialogElement;
 globalThis.history={state:{},pushState(value:unknown){this.state=value;},back(){this.state={};window.dispatchEvent(new Event('popstate'));}} as History;
 try{const stop=new BrowserReaderPlatform().panel(new FakeDialog(),()=>closed++);history.back();expect(closed).toBe(1);expect(listeners).toBe(0);stop();expect(listeners).toBe(0);}finally{globalThis.window=oldWindow;globalThis.history=oldHistory;globalThis.HTMLDialogElement=oldDialog;}
});

test('reader opt-outs survive persisted preference round-trip with safe existing defaults',()=>{
 const original=globalThis.localStorage;const values=new Map<string,string>();
 globalThis.localStorage={getItem:(key:string)=>values.get(key)||null,setItem:(key:string,value:string)=>{values.set(key,value);}} as Storage;
 try{const platform=new BrowserReaderPlatform();expect(platform.preferences('manga').saveProgress).toBe(true);expect(platform.preferences('manga').prefetchNext).toBe(true);platform.savePreferences({...platform.preferences('manga'),saveProgress:false,prefetchNext:false});expect(platform.preferences('manga').saveProgress).toBe(false);expect(platform.preferences('manga').prefetchNext).toBe(false);}finally{globalThis.localStorage=original;}
});
