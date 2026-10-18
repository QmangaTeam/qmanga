export type CommentSegment={kind:'text'|'bold'|'italic'|'strike';text:string};

// This small Markdown subset is rendered as Svelte text nodes, never as HTML.
export function commentSegments(body:string):CommentSegment[]{
 const segments:CommentSegment[]=[];
 const pattern=/\*\*([^*\n]+)\*\*|~~([^~\n]+)~~|\*([^*\n]+)\*/g;
 let position=0;
 for(const match of body.matchAll(pattern)){
  const index=match.index??0;
  if(index>position)segments.push({kind:'text',text:body.slice(position,index)});
  segments.push({kind:match[1]!==undefined?'bold':match[2]!==undefined?'strike':'italic',text:match[1]??match[2]??match[3]});
  position=index+match[0].length;
 }
 if(position<body.length)segments.push({kind:'text',text:body.slice(position)});
 return segments;
}
