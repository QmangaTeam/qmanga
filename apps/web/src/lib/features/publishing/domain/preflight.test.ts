import {expect,test} from 'bun:test';
import {processingSummary,submissionObstacles} from './preflight';
const ready={state:'ready',pages:[{id:'one',status:'ready'}],source:'Licensed source',credits:'Team',rightsConfirmed:true};
test('processing counts use every real page and preserve failed IDs',()=>{
 expect(processingSummary([{id:'a',status:'ready'},{id:'b',status:'queued'},{id:'c',status:'failed'},{id:'d',status:'uploading'}])).toEqual({ready:1,processing:2,failed:1,unknown:0,failedIDs:['c']});
});
test('preflight reflects saved details, transfer, page order and actual permission',()=>{
 expect(submissionObstacles(ready,{editor:true})).toEqual([]);
 expect(submissionObstacles({...ready,pages:[]},{editor:true})).toContain('pages');
 expect(submissionObstacles({...ready,state:'processing',pages:[{id:'one',status:'failed'}]},{editor:false,orderDirty:true,detailsDirty:true,uploading:true})).toEqual(['permission','transfer','failed','processing','order','details']);
 expect(submissionObstacles({...ready,source:'short',rightsConfirmed:false},{editor:true})).toContain('details');
});
