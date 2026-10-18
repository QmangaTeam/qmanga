import {test,expect} from 'bun:test';
import {validNickname,credentialError} from './account';
test('registration requires a safe nickname while legacy login remains available',()=>{
 for(const name of ['Оқырман','Қазақ оқырманы','Reader_42'])expect(validNickname(name)).toBe(true);
 for(const name of ['','x',' leading','two  spaces','reader@example.com','emoji😀'])expect(validNickname(name)).toBe(false);
 expect(credentialError({email:'a@example.com',password:'long-password'},true)).toBe('nickname');
 expect(credentialError({email:'a@example.com',password:'long-password'},false)).toBe(null);
});
