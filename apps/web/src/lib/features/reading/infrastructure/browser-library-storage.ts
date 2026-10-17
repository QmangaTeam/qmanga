import type {LibraryLocal,LibraryStorage} from '../application/library-sync';
import {AppFailure} from '$lib/shared/domain/failure';
function database():Promise<IDBDatabase>{return new Promise((resolve,reject)=>{const request=indexedDB.open('qmanga-library',1);request.onupgradeneeded=()=>{const store=request.result.createObjectStore('selections',{keyPath:['account','selection.titleId']});store.createIndex('account','account');};request.onsuccess=()=>resolve(request.result);request.onerror=()=>reject(new AppFailure('storage_unavailable'));});}
async function execute<Value>(mode:IDBTransactionMode,operation:(store:IDBObjectStore)=>IDBRequest<Value>):Promise<Value>{const connection=await database();return new Promise((resolve,reject)=>{const transaction=connection.transaction('selections',mode),request=operation(transaction.objectStore('selections'));let value:Value;request.onsuccess=()=>value=request.result;transaction.oncomplete=()=>{connection.close();resolve(value);};transaction.onerror=transaction.onabort=()=>{connection.close();reject(new AppFailure('storage_unavailable'));};});}
export class BrowserLibraryStorage implements LibraryStorage {
 private epoch:string|undefined;
 private check(){const current=localStorage.getItem('qmanga:auth-epoch')||'';if(this.epoch===undefined)this.epoch=current;if(this.epoch!==current)throw new AppFailure('account_changed');}
 async read(account:string,title:string){this.check();return(await execute<LibraryLocal|undefined>('readonly',store=>store.get([account,title])))||null;}
 async write(row:LibraryLocal){this.check();await execute('readwrite',store=>{this.check();return store.put(row);});}
 async list(account:string){this.check();return execute<LibraryLocal[]>('readonly',store=>store.index('account').getAll(account));}
 async clear(){await execute('readwrite',store=>store.clear());}
}
