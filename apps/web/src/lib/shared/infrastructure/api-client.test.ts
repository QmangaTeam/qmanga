import {test,expect} from 'bun:test';
import {ApiClient} from './api-client';
test('transport errors do not expose English network or HTML parser messages',async()=>{
 const failed=new ApiClient((async()=>{throw new Error('Failed to fetch');}) as typeof fetch);
 await expect(failed.call('/api/v1/titles')).rejects.toThrow('Қосылым');
 const html=new ApiClient((async()=>new Response('<html>unavailable</html>',{status:503})) as typeof fetch);
 await expect(html.call('/api/v1/titles')).rejects.toThrow('Қызмет');
});
