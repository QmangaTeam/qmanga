import {test,expect} from 'bun:test';
import {constraintMessage} from './form-validation';
import {allowDiscard} from './unsaved';
test('cancel preserves unsaved work and clean changes do not prompt',()=>{
 let calls=0;const ask=()=>{calls++;return false;};
 expect(allowDiscard(false,ask)).toBe(true);expect(calls).toBe(0);
 expect(allowDiscard(true,ask)).toBe(false);expect(calls).toBe(1);
 expect(allowDiscard(true,()=>true)).toBe(true);
});
test('native input constraints have localised field messages without classifying network failures',()=>{
 expect(constraintMessage({valueMissing:true},'ru')).toBe('Заполни это поле.');
 expect(constraintMessage({tooShort:true},'kk')).toBe('Мәтін тым қысқа.');
 expect(constraintMessage({typeMismatch:true},'ru')).toContain('формат');
 expect(constraintMessage({},'ru')).toBe('Проверь значение поля.');
});
