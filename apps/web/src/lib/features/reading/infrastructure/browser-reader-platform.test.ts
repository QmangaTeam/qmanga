import {test,expect} from 'bun:test';
import {BrowserReaderPlatform} from './browser-reader-platform';
test('Back closes the drawer and removes its history listener exactly once',()=>{
 const oldWindow=globalThis.window,oldHistory=globalThis.history,oldDialog=globalThis.HTMLDialogElement;let listeners=0,closed=0;
 class FakeWindow extends EventTarget{addEventListener(type:string,listener:EventListenerOrEventListenerObject,options?:AddEventListenerOptions|boolean){if(type==='popstate')listeners++;super.addEventListener(type,listener,options);}removeEventListener(type:string,listener:EventListenerOrEventListenerObject,options?:EventListenerOptions|boolean){if(type==='popstate')listeners--;super.removeEventListener(type,listener,options);}}
 class FakeDialog extends EventTarget{showModal(){}close(){}}
 globalThis.window=new FakeWindow() as unknown as Window & typeof globalThis;globalThis.HTMLDialogElement=FakeDialog as unknown as typeof HTMLDialogElement;
 globalThis.history={state:{},pushState(value:unknown){this.state=value;},back(){this.state={};window.dispatchEvent(new Event('popstate'));}} as History;
 try{const stop=new BrowserReaderPlatform().panel(new FakeDialog(),()=>closed++);history.back();expect(closed).toBe(1);expect(listeners).toBe(0);stop();expect(listeners).toBe(0);}finally{globalThis.window=oldWindow;globalThis.history=oldHistory;globalThis.HTMLDialogElement=oldDialog;}
});
