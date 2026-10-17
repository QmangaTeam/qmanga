export function movePage(ids:string[],id:string,position:number):string[]{
 if(!Number.isInteger(position)||position<1||position>ids.length||!ids.includes(id))return [...ids];
 const result=ids.filter(item=>item!==id);result.splice(position-1,0,id);return result;
}
