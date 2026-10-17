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
