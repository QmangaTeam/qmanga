import {test,expect} from 'bun:test';
import {emptyPublication,validPublication} from './publication';
test('publication accepts optional data but rejects executable links and credentials',()=>{
 expect(validPublication(emptyPublication())).toBe(true);
 expect(validPublication({...emptyPublication(),year:2023,links:[{label:'Official',url:'https://example.com/title'}]})).toBe(true);
 for(const url of ['javascript:alert(1)','data:text/html,test','https://user:password@example.com','/relative'])expect(validPublication({...emptyPublication(),links:[{label:'Official',url}]})).toBe(false);
 expect(validPublication({...emptyPublication(),year:99999})).toBe(false);
});
