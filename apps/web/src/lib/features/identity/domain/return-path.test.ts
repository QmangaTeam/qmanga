import {expect,test} from 'bun:test';
import {safeReturnPath} from './return-path';
test('login returns to a local reader or filtered catalogue, never an external or auth URL',()=>{
 expect(safeReturnPath('/read/release?mode=pages#end')).toBe('/read/release?mode=pages#end');
 expect(safeReturnPath('/catalog?q=%D2%9B')).toBe('/catalog?q=%D2%9B');
 for(const value of [null,'https://evil.test','//evil.test','/\\evil.test','/login?returnTo=/login','/api/v1/auth/logout','/auth/complete?returnTo=/auth/complete','/\n/evil.test'])expect(safeReturnPath(value)).toBe('/');
});
