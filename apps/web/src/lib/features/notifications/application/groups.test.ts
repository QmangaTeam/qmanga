import {test,expect} from 'bun:test';
import {notificationGroups} from './groups';
import type {Notification} from './notifications';
test('chapter notifications group by title without merging moderation or replies',()=>{
 const items=[{id:'one',kind:'new_release',titleSlug:'title'},{id:'two',kind:'comment_reply',titleSlug:'title'},{id:'three',kind:'new_release',titleSlug:'title'},{id:'four',kind:'new_release',titleSlug:'other'}] as Notification[];
 expect(notificationGroups(items).map(group=>group.map(item=>item.id))).toEqual([['one','three'],['two'],['four']]);
});
