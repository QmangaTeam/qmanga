import type {Position} from '../domain/policies';
import type {ReadingSession} from './progress';

export async function initializeReaderPosition(linked:Position|null,resume:()=>Promise<Position|null>,ports:{go(position:Position):Promise<void>;userNavigated():boolean;ready():void;quality():Promise<void>}) {
 if(linked)await ports.go(linked);
 const saved=await resume();
 if(!linked&&!ports.userNavigated()&&saved)await ports.go(saved);
 ports.ready();
 await ports.quality();
}

export async function resolveReaderProgress(session:ReadingSession,useLocal:boolean,currentRelease:string,ports:{go(position:Position):Promise<void>;openRelease(id:string):void}) {
 const position=await session.resolve(useLocal);
 if(useLocal||!position)return;
 if(position.releaseId!==currentRelease){session.close();ports.openRelease(position.releaseId);}
 else await ports.go(position);
}
