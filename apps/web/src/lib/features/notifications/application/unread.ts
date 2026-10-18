export function unreadBadge(accountId:string|null,count:number):string{
 return accountId&&Number.isSafeInteger(count)&&count>0?(count>99?'99+':String(count)):'';
}
