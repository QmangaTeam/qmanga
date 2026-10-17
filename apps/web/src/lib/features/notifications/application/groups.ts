import type {Notification} from './notifications';
export function notificationGroups(items:Notification[]):Notification[][]{
 const groups=new Map<string,Notification[]>();
 for(const item of items){const key=item.kind==='new_release'&&item.titleSlug?'title:'+item.titleSlug:'item:'+item.id;const group=groups.get(key)||[];group.push(item);groups.set(key,group);}
 return [...groups.values()];
}
