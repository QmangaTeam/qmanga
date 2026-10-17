import {test,expect} from 'bun:test';
import {downloadFrom} from './download-record';
import type {Download} from '../application/downloads';
test('legacy downloads retain identity files bytes and manifest when notices are introduced',()=>{
 const legacy={id:'release:revision:eco',status:'ready',files:['/api/v1/media/file'],bytes:123,total:123,message:'old message',manifest:{releaseId:'release'},tier:'eco',updatedAt:4} as Download;
 const restored=downloadFrom(legacy);
 expect(restored.notice).toBe('none');
 expect(restored.id).toBe(legacy.id);
 expect(restored.files).toEqual(legacy.files);
 expect(restored.manifest).toBe(legacy.manifest);
 expect(restored.bytes).toBe(123);
 expect(downloadFrom({...legacy,status:'partial'}).notice).toBe('download_partial');
 expect(downloadFrom({...legacy,status:'error'}).notice).toBe('download_failed');
});
