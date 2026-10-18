import {test,expect} from 'bun:test';
import {render} from './markup';
import TeamFixture from './Team.fixture.svelte';
test('solo portfolio exposes an explicit Telegram link and its existing project',()=>{
 const team={id:'team',name:'Дала тобы',slug:'dala',description:'Қазақша аудармалар',telegramURL:'https://t.me/dala_team',members:[{name:'Dala',role:'owner'}],projects:[{id:'project',titleSlug:'story',titleName:'Story'}]};
 const russian=render(TeamFixture,{props:{language:'ru',team}}).body;
 expect(russian).toContain('href="https://t.me/dala_team"');expect(russian).toContain('Один переводчик');expect(russian).toContain('href="/titles/story"');
 const kazakh=render(TeamFixture,{props:{language:'kk',team}}).body;
 expect(kazakh).toContain('Жеке аудармашы');expect(kazakh).toContain('ДТ');
 expect(kazakh).toContain('src="/media/avatars/team/team"');expect(kazakh).toContain('alt=""');expect(kazakh).toContain('aria-hidden="true"');
});
