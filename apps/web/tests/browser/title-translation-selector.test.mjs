import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {chromium} from 'playwright';
const fixture=JSON.parse(await readFile(process.env.QMANGA_FIXTURE_FILE,'utf8'));
assert.ok(['localhost','127.0.0.1'].includes(new URL(fixture.app_url).hostname));
test('title translation selection retains exact editions through sorting and reset',{timeout:45000},async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.QMANGA_BROWSER_EXECUTABLE?{executablePath:process.env.QMANGA_BROWSER_EXECUTABLE}:{})});
 try{const context=await browser.newContext({baseURL:fixture.app_url,viewport:{width:320,height:807}});
  await context.addCookies([{name:'qmanga_locale',value:'ru',url:fixture.app_url}]);
  const detail=await(await context.request.get('/api/v1/titles/'+fixture.title_slug)).json();
  const team=detail.projects[0];assert.ok(team);
  const page=await context.newPage();
  await page.goto('/titles/'+fixture.title_slug+'?team='+encodeURIComponent(team.team_slug));
  await page.locator('.translation-team[aria-current]').waitFor();
  assert.equal(await page.locator('.translation-team[aria-current]').innerText(),team.team_name);
  const result=await(await context.request.get('/api/v1/titles/'+fixture.title_slug+'/chapters?team='+encodeURIComponent(team.team_slug))).json();
  assert.ok(result.items.length>0);
  for(const chapter of result.items){assert.equal(chapter.releases.length,1);assert.equal(chapter.releases[0].team_id,team.team_id);}
  assert.equal(await page.locator('.edition-picker').count(),0);
  const first=result.items[0].releases[0];assert.equal(await page.locator('.read-button').getAttribute('href'),'/read/'+first.id);
  await Promise.all([page.waitForURL(u=>u.searchParams.get('sort')==='desc'),page.locator('.chapter-search button[name="sort"]').last().click()]);
  assert.equal(new URL(page.url()).searchParams.get('team'),team.team_slug);
  assert.equal(await page.locator('input[name="team"]').inputValue(),team.team_slug);
  assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+1));
  await Promise.all([page.waitForURL(u=>!u.searchParams.has('team')),page.locator('.all-translations').click()]);
  assert.equal(new URL(page.url()).searchParams.get('sort'),'desc');
  await context.close();
 }finally{await browser.close();}
});
