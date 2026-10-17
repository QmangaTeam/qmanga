import {test,expect} from 'bun:test';
import {ThrottledSave} from './throttled-save';
test('continuous scroll does not postpone persistence and closing flushes the latest position',async()=>{
 let callback:()=>void=()=>{},scheduled=0;const saved:number[]=[];
 const saver=new ThrottledSave<number>({later:action=>{callback=action;scheduled++;return ()=>{};}},async value=>{saved.push(value);});
 saver.offer(1);saver.offer(2);saver.offer(3);expect(scheduled).toBe(1);callback();await saver.flush();expect(saved).toEqual([3]);
 saver.offer(4);await saver.flush();expect(saved).toEqual([3,4]);
});
