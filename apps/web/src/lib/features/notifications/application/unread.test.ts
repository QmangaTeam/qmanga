import {test,expect} from 'bun:test';
import {unreadBadge} from './unread';
test('badge hides zero and guests and caps visible text without losing real accessible count',()=>{
 expect(unreadBadge(null,101)).toBe('');expect(unreadBadge('user',0)).toBe('');
 expect(unreadBadge('user',8)).toBe('8');expect(unreadBadge('user',100)).toBe('99+');
 expect(unreadBadge('user',-1)).toBe('');expect(unreadBadge('user',NaN)).toBe('');
});
