export const workspaceSteps=['draft-pages','draft-details','draft-preview'] as const;
export function boundedPosition(value:number){return Number.isFinite(value)?Math.max(0,Math.min(1000000,Math.round(value))):0;}

function validProjectList(value:string):boolean{
 if(value==='/studio/projects')return true;
 if(!value.startsWith('/studio/projects?')||!/^\/studio\/projects\?[A-Za-z0-9_%=&+.*-]*$/.test(value))return false;
 const seen=new Set<string>();
 try{
  for(const part of value.slice('/studio/projects?'.length).split('&')){
   const separator=part.indexOf('=');if(separator<0)return false;
   const key=part.slice(0,separator),field=decodeURIComponent(part.slice(separator+1).replace(/\+/g,' '));
   if(seen.has(key))return false;seen.add(key);
   if(key==='q'){if([...field].length>80)return false;}
   else if(key==='team'){if(!/^[A-Za-z0-9_-]{0,64}$/.test(field))return false;}
   else if(key==='pages'){if(!/^\d{1,2}$/.test(field)||Number(field)<1||Number(field)>20)return false;}
   else if(key==='position'){if(!/^\d{1,7}$/.test(field)||Number(field)>1000000)return false;}
   else return false;
  }
 }catch{return false;}
 return true;
}
export function projectListReturn(value:string|null|undefined){return value&&validProjectList(value)?value:'/studio/projects';}

export function workspaceLocation(project:string,revision:string,step:string,position=0,projectsReturn?:string|null){
 const section=workspaceSteps.includes(step as typeof workspaceSteps[number])?step:'draft-preview';
 const list=projectsReturn?projectListReturn(projectsReturn):'';
 return '/studio/projects/'+encodeURIComponent(project)+'?revision='+encodeURIComponent(revision)+'&step='+section+'&position='+boundedPosition(position)+(list?'&projectsReturn='+encodeURIComponent(list):'');
}
// Accept only the workspace shape emitted above. A nested list contains criteria, never an account scope.
export function workspaceReturn(value:string|null|undefined,revision:string){
 if(!value)return '/studio/projects';
 const match=/^\/studio\/projects\/([A-Za-z0-9_-]+)\?revision=([A-Za-z0-9_-]+)&step=(draft-pages|draft-details|draft-preview)&position=(\d{1,7})(?:&projectsReturn=([^&]+))?$/.exec(value);
 if(!match||match[2]!==revision||Number(match[4])>1000000)return '/studio/projects';
 if(match[5]){
  try{const list=decodeURIComponent(match[5]);if(!validProjectList(list)||encodeURIComponent(list)!==match[5])return '/studio/projects';}catch{return '/studio/projects';}
 }
 return value;
}
