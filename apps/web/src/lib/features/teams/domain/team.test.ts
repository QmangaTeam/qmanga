import {test,expect} from 'bun:test';
import {validTeam} from './team';
test('public team links accept only a public HTTPS Telegram username',()=>{
 const input={slug:'dala',name:'Дала',description:'',telegramURL:'https://t.me/dala_team'};
 expect(validTeam(input)).toBe(true);
 for(const telegramURL of ['javascript:alert(1)','https://t.me@evil.test/dala_team','https://t.me/+private','https://t.me/dala_team?start=secret'])expect(validTeam({...input,telegramURL})).toBe(false);
});
