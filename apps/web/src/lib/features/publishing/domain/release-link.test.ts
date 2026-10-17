import {test,expect} from 'bun:test';
import {publicationReleaseId} from './release-link';
test('withdrawal selects a public reader link on this site only',()=>{
 expect(publicationReleaseId('/read/release-1','http://localhost:8080')).toBe('release-1');
 expect(publicationReleaseId('http://localhost:8080/read/release-1#page=2','http://localhost:8080')).toBe('release-1');
 for(const value of ['https://another.test/read/release-1','/studio/projects/project','/read/release-1?preview=draft','/read/','/read/one/extra','javascript:alert(1)'])expect(publicationReleaseId(value,'http://localhost:8080')).toBeNull();
});
