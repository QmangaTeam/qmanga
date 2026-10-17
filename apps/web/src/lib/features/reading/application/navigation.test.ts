import {expect,test} from 'bun:test';
import {PageNavigation,readingDirection,pageStep} from './navigation';

test('strip controls advance on the right even for an RTL title',()=>{
 expect(readingDirection('strip','auto','rtl')).toBe('ltr');
 expect(readingDirection('pages','auto','rtl')).toBe('rtl');
 expect(readingDirection('pages','ltr','rtl')).toBe('ltr');
 expect(pageStep('right','ltr')).toBe(1);
 expect(pageStep('left','ltr')).toBe(-1);
 expect(pageStep('right','rtl')).toBe(-1);
 expect(pageStep('left','rtl')).toBe(1);
});

test('rapid inputs advance from the requested page, not a stale scroll observation',()=>{
 const navigation=new PageNavigation(5);
 const first=navigation.move(1)!;
 navigation.observe(0);
 const second=navigation.move(1)!;
 expect(navigation.index).toBe(2);
 expect(navigation.current(first)).toBe(false);
 expect(navigation.current(second)).toBe(true);
 navigation.finish(first);
 navigation.observe(0);
 expect(navigation.index).toBe(2);
 navigation.finish(second);
 navigation.observe(3);
 expect(navigation.index).toBe(3);
});

test('page limits and explicit selection share the same cursor',()=>{
 const navigation=new PageNavigation(5);
 expect(navigation.move(-1)).toBeNull();
 const request=navigation.select(4)!;
 expect(navigation.index).toBe(4);
 expect(navigation.move(1)).toBeNull();
 expect(navigation.current(request)).toBe(true);
 expect(navigation.select(9)).toBeNull();
 expect(navigation.select(-1)).toBeNull();
 expect(navigation.select(1.5)).toBeNull();
 navigation.finish(request);
 navigation.observe(2);
 expect(navigation.index).toBe(2);
 expect(navigation.move(1)?.index).toBe(3);
});
