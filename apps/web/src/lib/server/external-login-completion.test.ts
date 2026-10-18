import {test,expect} from 'bun:test';
import {load} from '../../routes/auth/complete/+page.server';
const nonce='a'.repeat(43),session='synthetic-session-B';
async function marker(token=session){const bytes=new Uint8Array(await crypto.subtle.digest('SHA-256',new TextEncoder().encode(nonce+'\0'+token)));return [...bytes].map(value=>value.toString(16).padStart(2,'0')).join('');}
function request(proof:string|undefined,user=true,path='/library'){
 const values=new Map([['qmanga_session',session]]);if(proof!==undefined)values.set('qmanga_external_complete',proof);let consumed=false;
 const input={url:new URL('https://app.test/auth/complete?'+new URLSearchParams({nonce,returnTo:path})),cookies:{get:(name:string)=>values.get(name),delete:(name:string)=>{consumed=true;values.delete(name);}},parent:async()=>({user:user?{id:'B'}:null})};
 return{input:input as unknown as Parameters<typeof load>[0],values,consumed:()=>consumed};
}
test('valid callback proof is session-bound, consumed once and exposes only safe destination',async()=>{const r=request(await marker());expect(await load(r.input)).toEqual({returnTo:'/library'});expect(r.consumed()).toBe(true);await expect(load(r.input)).rejects.toMatchObject({status:303,location:'/library'});});
test('unmarked, wrong-session and unauthenticated landings redirect without allowing browser cleanup',async()=>{for(const r of [request(undefined),request(await marker('previous-session-A')),request(await marker(),false)]){await expect(load(r.input)).rejects.toMatchObject({status:303,location:'/library'});expect(r.consumed()).toBe(true);}});
test('a valid proof cannot loop to the completion route or redirect off-origin',async()=>{for(const path of ['/auth/complete?returnTo=/auth/complete','//evil.example'])expect(await load(request(await marker(),true,path).input)).toEqual({returnTo:'/'});});
