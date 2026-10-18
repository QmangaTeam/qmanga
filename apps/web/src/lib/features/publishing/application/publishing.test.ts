import {test,expect} from 'bun:test';
import {UploadQueue,replacePage,type PublishingGateway,type Revision,type UploadSource,type StudioPage} from './publishing';
const page=(id:string,position:number,status='ready',uploadKey=id):StudioPage=>({id,position,status,uploadKey,width:10,height:20,bytes:50,failure:''});
const revision=(pages:StudioPage[]):Revision=>({id:'revision',releaseId:'release',projectId:'project',chapterId:'chapter',titleId:'title',titleName:'Title',chapterLabel:'1',teamName:'Team',state:'ready',note:'',version:4,pages});
const source=(name:string):UploadSource=>({name,key:name,size:50,content:{}});

test('resume skips stored pages, retries failed slots and keeps natural order',async()=>{
 const calls:{key:string;position:number}[]=[];
 const gateway={upload:(_id:string,file:UploadSource,position:number)=>{calls.push({key:file.key,position});return {done:Promise.resolve(),cancel(){}};}} as unknown as PublishingGateway;
 const queue=new UploadQueue(gateway);
 await queue.run(revision([page('one',1,'ready','1.png'),page('two',2,'failed','2.png')]),[source('10.png'),source('1.png'),source('2.png')],()=>{});
 expect(calls).toEqual([{key:'2.png',position:2},{key:'10.png',position:3}]);
});

test('failed upload never deletes the page being replaced',async()=>{
 let reordered=false;
 const gateway={upload:()=>({done:Promise.reject(new Error('offline')),cancel(){}}),reorder:async()=>{reordered=true;}} as unknown as PublishingGateway;
 await expect(replacePage(gateway,revision([page('old',1)]),'old',source('new.png'),()=>{})).rejects.toThrow('offline');
 expect(reordered).toBe(false);
});

test('replacement changes one slot with the latest CAS version and preserves other pages',async()=>{
 let current=revision([page('first',1),page('old',2),page('last',3)]);
 let order:string[]=[];let version=0;
 const gateway={
  upload:(_id:string,file:UploadSource,position:number)=>{current={...current,version:6,pages:[...current.pages,page('new',position,'queued',file.key)]};return {done:Promise.resolve(),cancel(){}};},
  revision:async()=>current,
  reorder:async(_id:string,v:number,ids:string[])=>{version=v;order=ids;current={...current,version:v+1,pages:ids.map((id,index)=>({...current.pages.find(p=>p.id===id)!,position:index+1}))};}
 } as unknown as PublishingGateway;
 const result=await replacePage(gateway,current,'old',source('new.png'),()=>{});
 expect(order).toEqual(['first','new','last']);expect(version).toBe(6);expect(result.pages).toHaveLength(3);
});

test('replacement refuses to remove anything if the selected page was removed concurrently',async()=>{
 let reordered=false;
 const gateway={upload:()=>({done:Promise.resolve(),cancel(){}}),revision:async()=>revision([page('new',1,'queued','replace:old:new.png')]),reorder:async()=>{reordered=true;}} as unknown as PublishingGateway;
 await expect(replacePage(gateway,revision([page('old',1)]),'old',source('new.png'),()=>{})).rejects.toBeDefined();
 expect(reordered).toBe(false);
});
