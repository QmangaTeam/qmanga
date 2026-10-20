import {expect,test} from 'bun:test';
import {publicationStates} from './messages';
test('draft label is distinct from views and project terminology in Kazakh',()=>{
 expect(publicationStates('kk').draft).toBe('Нобай');expect(publicationStates('ru').draft).toBe('Черновик');
 expect(publicationStates('kk').draft).not.toBe('Қаралым');
});
