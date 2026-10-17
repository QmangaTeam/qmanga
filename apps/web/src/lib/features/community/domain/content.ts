export function validComment(value:string):boolean {const length=[...value.trim()].length;return length>0&&length<=2000;}
export function validReport(value:string):boolean {const length=[...value.trim()].length;return length>=8&&length<=2000;}
export const categoryNames:Record<string,string>={image:'Сурет қатесі',translation:'Аударма қатесі',spoiler:'Спойлер',abuse:'Әдепсіз мазмұн',rights:'Құқық мәселесі',other:'Басқа мәселе'};
