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
