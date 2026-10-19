import {test,expect} from 'bun:test';
import {ApiClient} from './api-client';
test('transport errors do not expose English network or HTML parser messages',async()=>{
 const failed=new ApiClient((async()=>{throw new Error('Failed to fetch');}) as typeof fetch);
 await expect(failed.call('/api/v1/titles')).rejects.toThrow('Қосылым');
 const html=new ApiClient((async()=>new Response('<html>unavailable</html>',{status:503})) as typeof fetch);
 await expect(html.call('/api/v1/titles')).rejects.toThrow('Қызмет');
});
test('a long-lived client obtains a new CSRF token after its protected account changes',async()=>{
 let tokens=0;const sent:string[]=[];
 const api=new ApiClient((async(path:string,init?:RequestInit)=>{if(path.endsWith('/auth/csrf'))return Response.json({csrf_token:'token-'+(++tokens)});if(init?.method==='POST'){sent.push(new Headers(init.headers).get('X-CSRF-Token')||'');return new Response(null,{status:204});}return Response.json({allowed:true});}) as typeof fetch);
 await api.call('/api/v1/auth/guest-import','GET',undefined,'first');await api.call('/api/v1/auth/guest-import/complete','POST',{},'first');
 await api.call('/api/v1/auth/guest-import','GET',undefined,'second');await api.call('/api/v1/auth/guest-import/complete','POST',{},'second');
 expect(sent).toEqual(['token-1','token-2']);
});
