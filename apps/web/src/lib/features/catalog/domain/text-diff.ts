export type TextFragment={text:string;changed:boolean};
/** Linear, bounded comparison: unchanged prefix/suffix stay readable even for 10k descriptions. */
export function textDiff(before:string,after:string):{before:TextFragment[];after:TextFragment[]}{
 let start=0,end=0;while(start<before.length&&start<after.length&&before[start]===after[start])start++;
 while(end<before.length-start&&end<after.length-start&&before[before.length-1-end]===after[after.length-1-end])end++;
 while(start>0&&/[^\s]/u.test(before[start-1]||'')&&/[^\s]/u.test(before[start]||''))start--;
 while(end>0&&/[^\s]/u.test(before[before.length-end-1]||'')&&/[^\s]/u.test(before[before.length-end]||''))end--;
 const fragments=(value:string):TextFragment[]=>[{text:value.slice(0,start),changed:false},{text:value.slice(start,value.length-end),changed:true},{text:end?value.slice(-end):'',changed:false}].filter(part=>part.text.length>0);
 return {before:fragments(before),after:fragments(after)};
}
