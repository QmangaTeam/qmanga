export interface Publication {year:number;publishers:string[];source:string;country:string;formats:string[];links:{label:string;url:string}[]}
export function emptyPublication():Publication{return {year:0,publishers:[],source:'',country:'',formats:[],links:[]};}
export function safeExternalURL(value:string):boolean {
 try {const url=new URL(value);return ['https:','http:'].includes(url.protocol)&&!!url.hostname&&!url.username&&!url.password;}catch{return false;}
}
export function validPublication(value:Publication):boolean {
 return Number.isInteger(value.year)&&(value.year===0||(value.year>=1000&&value.year<=2200))&&value.publishers.length<=12&&value.formats.length<=8&&value.links.length<=12
 &&[value.source,value.country,...value.publishers,...value.formats].every(text=>[...text].length<=160&&!/[\r\n]/.test(text))
 &&value.links.every(link=>!!link.label.trim()&&[...link.label].length<=80&&!/[\r\n]/.test(link.label)&&link.url.length<=2000&&safeExternalURL(link.url));
}
