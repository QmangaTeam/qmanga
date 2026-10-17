import { expect, test } from 'bun:test';
import { anchorOffset, layout, visible, type Manifest } from './reader';
const manifest = { pages: [{ id:'page',position:1,variants:{eco:[{id:'first',width:500,height:1000},{id:'second',width:500,height:500}],high:[{id:'high',width:1000,height:3000}]}}]} as unknown as Manifest;
test('the source anchor survives fragment and quality changes', () => {
 const small=layout(manifest,'eco','strip',0),large=layout(manifest,'high','strip',0);
 const current=visible(small.items,{width:360,height:100,top:900});
 expect(current.position.fraction).toBeCloseTo(5/6);
 expect(anchorOffset(large.items,current.position,360)).toBeCloseTo(900);
 expect(small.total).toBe(3);
});

test('browser rounding at a page boundary does not leave navigation on the previous page',()=>{
 const chapter={pages:[{id:'first-page',variants:{eco:[{id:'first',width:720,height:1018}]}},{id:'next-page',variants:{eco:[{id:'next',width:720,height:1018}]}}]} as unknown as Manifest;
 const sheet=layout(chapter,'eco','strip',0);
 for(const width of [320,360,424,1024]){
  const target=anchorOffset(sheet.items,{pageId:'next-page',fraction:0},width);
  const current=visible(sheet.items,{width,height:640,top:Math.floor(target)});
  expect(current.pageIndex).toBe(1);
  expect(current.position).toEqual({pageId:'next-page',fraction:0});
  expect(visible(sheet.items,{width,height:640,top:Math.floor(target)-2}).pageIndex).toBe(0);
 }
});
