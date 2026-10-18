import {expect,test} from 'bun:test';
import {commentSegments} from './comment-markup';

test('comment formatting remains a safe, bounded text subset',()=>{
 expect(commentSegments('A **bold** *italic* ~~strike~~ <script>alert(1)</script>')).toEqual([
  {kind:'text',text:'A '},{kind:'bold',text:'bold'},{kind:'text',text:' '},{kind:'italic',text:'italic'},
  {kind:'text',text:' '},{kind:'strike',text:'strike'},{kind:'text',text:' <script>alert(1)</script>'}
 ]);
 expect(commentSegments('plain\ntext')).toEqual([{kind:'text',text:'plain\ntext'}]);
});
