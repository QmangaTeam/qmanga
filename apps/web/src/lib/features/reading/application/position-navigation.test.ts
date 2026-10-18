import {test,expect} from 'bun:test';
import {initializeReaderPosition,resolveReaderProgress} from './position-navigation';
import {ReadingSession,type ProgressStorage,type ReadingGateway,type LocalProgress} from './progress';
import type {Position} from '../domain/policies';

test('linked page is selected before slow quality loading and never resets later navigation',async()=>{
 let finishQuality!:()=>void,current:Position={pageId:'first',fraction:0},ready=false,qualityStarted=false;
 const quality=new Promise<void>(resolve=>finishQuality=resolve);
 const startup=initializeReaderPosition({pageId:'linked',fraction:0},async()=>({pageId:'saved',fraction:.7}),{go:async position=>{current=position;},userNavigated:()=>false,ready:()=>{ready=true;},quality:async()=>{qualityStarted=true;await quality;}});
 await Promise.resolve();await Promise.resolve();await Promise.resolve();
 expect(current.pageId).toBe('linked');expect(ready).toBe(true);expect(qualityStarted).toBe(true);
 current={pageId:'user-choice',fraction:.2};finishQuality();await startup;
 expect(current.pageId).toBe('user-choice');
});

test('a delayed resume never overrides navigation made while account progress loads',async()=>{
 let finishResume!:(value:Position)=>void,current='first',navigated=false;
 const resume=new Promise<Position>(resolve=>finishResume=resolve);
 const startup=initializeReaderPosition(null,()=>resume,{go:async position=>{current=position.pageId;},userNavigated:()=>navigated,ready:()=>{},quality:async()=>{}});
 current='user-choice';navigated=true;finishResume({pageId:'saved',fraction:.8});await startup;
 expect(current).toBe('user-choice');
});

test('accepting cloud progress opens its release instead of seeking a foreign page in the current one',async()=>{
 const session=new ReadingSession({} as ReadingGateway,{write:async()=>{}} as ProgressStorage,{later:()=>()=>{}},()=>{});
 session.row={scope:'account',progress:{releaseId:'old'},pending:null,conflict:{releaseId:'chosen',pageId:'chosen-page',fraction:.5,version:2}} as LocalProgress;
 const opened:string[]=[],pages:string[]=[];
 await resolveReaderProgress(session,false,'old',{go:async position=>{pages.push(position.pageId);},openRelease:id=>{opened.push(id);}});
 expect(opened).toEqual(['chosen']);expect(pages).toEqual([]);
 expect(session.row?.progress.pageId).toBe('chosen-page');
});
