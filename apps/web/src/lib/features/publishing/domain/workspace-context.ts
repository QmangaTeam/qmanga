export const workspaceSteps=['draft-pages','draft-details','draft-preview'] as const;
export function workspaceLocation(project:string,revision:string,step:string,position=0){
 const section=workspaceSteps.includes(step as typeof workspaceSteps[number])?step:'draft-preview';
 return '/studio/projects/'+encodeURIComponent(project)+'?revision='+encodeURIComponent(revision)+'&step='+section+'&position='+Math.max(0,Math.min(1000000,Math.round(position)));
}
// Deliberately accept only the exact shape emitted by the workspace, never an arbitrary return URL.
export function workspaceReturn(value:string|null|undefined,revision:string){
 if(!value)return '/studio/projects';
 const match=/^\/studio\/projects\/([A-Za-z0-9_-]+)\?revision=([A-Za-z0-9_-]+)&step=(draft-pages|draft-details|draft-preview)&position=(\d{1,7})$/.exec(value);
 if(!match||match[2]!==revision||Number(match[4])>1000000)return '/studio/projects';
 return value;
}

export function projectListReturn(value:string|null|undefined){return value&&/^\/studio\/projects(?:\?[A-Za-z0-9_%=&+.-]*)?$/.test(value)?value:'/studio/projects';}
