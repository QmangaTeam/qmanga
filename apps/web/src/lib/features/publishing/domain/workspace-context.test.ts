import {expect,test} from 'bun:test';
import {workspaceReturn,workspaceLocation} from './workspace-context';
test('preview return accepts only a local workspace for the exact preview revision',()=>{
 const href=workspaceLocation('project','revision','draft-preview',432);
 expect(workspaceReturn(href,'revision')).toBe(href);
 for(const value of ['https://evil.test/studio/projects/p?revision=revision','//evil.test','/studio/projects/p?revision=other','/studio/projects/p/metadata?revision=revision'])expect(workspaceReturn(value,'revision')).toBe('/studio/projects');
});
test('list return is limited to the project index',async()=>{
 const {projectListReturn}=await import('./workspace-context');
 expect(projectListReturn('/studio/projects?q=title&team=a&pages=2&position=400')).toBe('/studio/projects?q=title&team=a&pages=2&position=400');
 for(const href of ['//evil.test','/studio/projects/secret','/studio/projects?x="bad"','https://evil.test'])expect(projectListReturn(href)).toBe('/studio/projects');
});
test('preview round trip carries the validated applied list context through both returns',()=>{
 const list='/studio/projects?q=Applied+%D0%A2%D0%B0%D0%B9%D1%82%D0%BB&team=team-id&pages=2&position=600';
 const workspace=workspaceLocation('project','revision-b','draft-preview',320,list);
 expect(workspace).toContain('&projectsReturn='+encodeURIComponent(list));
 expect(workspaceReturn(workspace,'revision-b')).toBe(workspace);
 expect(decodeURIComponent(workspace.split('&projectsReturn=')[1])).toBe(list);
 expect(workspaceLocation('project','revision-b','draft-preview',320,'https://foreign.test')).not.toContain('foreign');
 for(const value of ['https://foreign.test','/studio/projects/private','/studio/projects?account=other','/studio/projects?pages=100000'])expect(workspaceReturn('/studio/projects/project?revision=revision-b&step=draft-preview&position=320&projectsReturn='+encodeURIComponent(value),'revision-b')).toBe('/studio/projects');
});
