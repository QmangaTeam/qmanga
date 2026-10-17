import {test,expect} from 'bun:test';
import {offlineComplete} from './policies';
test('saved readiness accepts canonical Go URLs, not external or traversal paths',()=>{
 const good={url:'/api/v1/media/asset-1',bytes:20};
 expect(offlineComplete('revision',[good],[{...good,revision:'revision'}])).toBe(true);
 for(const url of ['https://minio.example/media/asset','//evil.test/media/a','/api/v1/media/../source/a','/media/a?key=secret']){expect(offlineComplete('revision',[{...good,url}],[{...good,url,revision:'revision'}])).toBe(false);}
});
