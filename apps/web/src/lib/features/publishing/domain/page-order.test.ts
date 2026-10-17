import {test,expect} from 'bun:test';
import {movePage} from './page-order';
test('mobile position moves preserve every page and reject invalid destinations',()=>{
 expect(movePage(['one','two','three'],'three',1)).toEqual(['three','one','two']);
 expect(movePage(['one','two','three'],'one',3)).toEqual(['two','three','one']);
 expect(movePage(['one','two'],'one',0)).toEqual(['one','two']);
 expect(movePage(['one','two'],'missing',1)).toEqual(['one','two']);
});
