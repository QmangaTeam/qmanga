import {test,expect} from 'bun:test';
import {importGuest} from './guest-import';
import type {ReadingGateway,LocalProgress,ProgressStorage} from './progress';
test('guest import preserves favorite, team and notification choices already in the library',async()=>{
 let writes=0;const gateway={progress:async()=>({version:0}),advance:async()=>({outcome:'accepted'}),selection:async()=>({favorite:true,preferredTeamId:'preferred',notify:false}),save:async()=>{writes++;}} as unknown as ReadingGateway;
 const storage={identity:()=>({id:'event',deviceId:'device',sequence:1})} as ProgressStorage;
 await importGuest(gateway,storage,'account',{titleId:'title',progress:{releaseId:'release',revisionId:'rev',pageId:'page',fraction:.5}} as LocalProgress);
 expect(writes).toBe(0);
});
