import {AppFailure} from '../../../shared/domain/failure';
export type ImportOptions={signal?:AbortSignal;onProgress?:(done:number,total:number)=>void};
const MiB=1024*1024;
const invalid=()=>new AppFailure('import_invalid');
function cancelled(signal?:AbortSignal){if(signal?.aborted)throw new AppFailure('upload_cancelled');}
const crcTable=Uint32Array.from({length:256},(_,value)=>{let crc=value;for(let bit=0;bit<8;bit++)crc=(crc>>>1)^((crc&1)?0xedb88320:0);return crc>>>0;});
async function crc32(data:Uint8Array,signal?:AbortSignal){let crc=0xffffffff;for(let offset=0;offset<data.length;offset+=262144){cancelled(signal);for(let i=offset;i<Math.min(data.length,offset+262144);i++)crc=(crc>>>8)^crcTable[(crc^data[i])&255];await new Promise(resolve=>setTimeout(resolve,0));}return (crc^0xffffffff)>>>0;}
/** Reads ZIP entries into memory only. Nothing is written to a filesystem or uploaded before confirmation. */
export async function importCbz(file:File,options:ImportOptions={}):Promise<File[]>{
 const {signal,onProgress}=options;cancelled(signal);
 if(file.size===0||file.size>100*MiB)throw new AppFailure('import_size');
 const bytes=new Uint8Array(await file.arrayBuffer()),view=new DataView(bytes.buffer);cancelled(signal);
 let end=-1;for(let i=bytes.length-22;i>=Math.max(0,bytes.length-65557);i--)if(view.getUint32(i,true)===0x06054b50&&i+22+view.getUint16(i+20,true)===bytes.length){end=i;break;}
 if(end<0)throw invalid();
 const count=view.getUint16(end+10,true),centralSize=view.getUint32(end+12,true),centralOffset=view.getUint32(end+16,true);
 if(view.getUint16(end+4,true)||view.getUint16(end+6,true)||view.getUint16(end+8,true)!==count||count>1000||centralOffset+centralSize!==end)throw invalid();
 type Entry={name:string;method:number;flags:number;crc:number;packed:number;size:number;offset:number};
 const entries:Entry[]=[],seen=new Set<string>();let cursor=centralOffset,total=0;
 for(let i=0;i<count;i++){
  cancelled(signal);if(cursor+46>end||view.getUint32(cursor,true)!==0x02014b50)throw invalid();
  const flags=view.getUint16(cursor+8,true),method=view.getUint16(cursor+10,true),packed=view.getUint32(cursor+20,true),size=view.getUint32(cursor+24,true),nameSize=view.getUint16(cursor+28,true),extra=view.getUint16(cursor+30,true),comment=view.getUint16(cursor+32,true),attributes=view.getUint32(cursor+38,true),offset=view.getUint32(cursor+42,true),next=cursor+46+nameSize+extra+comment;
  if(next>end||view.getUint16(cursor+34,true)||(flags&~0x80e)!==0||![0,8].includes(method)||(attributes>>>16&0xf000)===0xa000||size===0xffffffff||packed===0xffffffff||offset===0xffffffff)throw invalid();
  let name:string;try{name=new TextDecoder('utf-8',{fatal:true}).decode(bytes.subarray(cursor+46,cursor+46+nameSize));}catch{throw invalid();}
  if(!name||name.includes('\\')||name.startsWith('/')||name.includes(':')||/[\x00-\x1f]/.test(name)||name.split('/').some(p=>p==='..'||p==='.')||seen.has(name))throw invalid();seen.add(name);
  total+=size;if(size>25*MiB||total>250*MiB||(size>MiB&&size>packed*200))throw new AppFailure('import_size');
  if(offset+30>centralOffset||view.getUint32(offset,true)!==0x04034b50||view.getUint16(offset+6,true)!==flags||view.getUint16(offset+8,true)!==method)throw invalid();
  const localName=view.getUint16(offset+26,true),localExtra=view.getUint16(offset+28,true),start=offset+30+localName+localExtra;
  if(start+packed>centralOffset||localName!==nameSize||!bytes.subarray(offset+30,offset+30+localName).every((byte,index)=>byte===bytes[cursor+46+index]))throw invalid();
  if(!(flags&8)&&(view.getUint32(offset+14,true)!==view.getUint32(cursor+16,true)||view.getUint32(offset+18,true)!==packed||view.getUint32(offset+22,true)!==size))throw invalid();
  if(!name.endsWith('/')&&!name.split('/').some(p=>p==='__MACOSX'||p.startsWith('.'))&&/\.(jpe?g|png|webp)$/i.test(name))entries.push({name,method,flags,crc:view.getUint32(cursor+16,true),packed,size,offset:start});
  cursor=next;
 }
 if(cursor!==end||!entries.length||entries.length>300)throw invalid();
 entries.sort((a,b)=>a.name.localeCompare(b.name,'en',{numeric:true,sensitivity:'variant'}));const files:File[]=[];
 for(const entry of entries){
  cancelled(signal);const packed=bytes.slice(entry.offset,entry.offset+entry.packed);let content:Uint8Array;
  if(entry.method===0){content=packed;}else{
   const reader=new Blob([packed]).stream().pipeThrough(new DecompressionStream('deflate-raw')).getReader();const chunks:Uint8Array[]=[];let length=0;
   const abort=()=>{void reader.cancel().catch(()=>{});};signal?.addEventListener('abort',abort,{once:true});
   try{while(true){cancelled(signal);const item=await reader.read();if(item.done)break;length+=item.value.length;if(length>entry.size||length>25*MiB)throw invalid();chunks.push(item.value);}content=new Uint8Array(length);let position=0;for(const chunk of chunks){content.set(chunk,position);position+=chunk.length;}}
   catch(failure){await reader.cancel().catch(()=>{});cancelled(signal);throw invalid();}finally{signal?.removeEventListener('abort',abort);reader.releaseLock();}
  }
  cancelled(signal);if(content.length!==entry.size||await crc32(content,signal)!==entry.crc)throw invalid();
  const ext=entry.name.split('.').pop()!.toLowerCase();files.push(new File([new Uint8Array(content).buffer],entry.name,{type:ext==='png'?'image/png':ext==='webp'?'image/webp':'image/jpeg',lastModified:0}));onProgress?.(files.length,entries.length);
  await new Promise(resolve=>setTimeout(resolve,0));
 }
 return files;
}
