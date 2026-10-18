import {test,expect} from 'bun:test';
import {BrowserDownloadStorage} from './browser-download-storage';

test('saving with an active worker does not wait for a waiting or failed update',async()=>{
 const descriptor=Object.getOwnPropertyDescriptor(globalThis,'navigator');let persisted=false;
 const update=new EventTarget() as EventTarget & {state:string};update.state='redundant';
 Object.defineProperty(globalThis,'navigator',{configurable:true,value:{serviceWorker:{register:async()=>({active:{state:'activated'},waiting:update}),ready:Promise.resolve({})},storage:{persist:async()=>{persisted=true;return true;}}}});
 try{await new BrowserDownloadStorage().install();expect(persisted).toBe(true);}finally{if(descriptor)Object.defineProperty(globalThis,'navigator',descriptor);else delete (globalThis as any).navigator;}
});
