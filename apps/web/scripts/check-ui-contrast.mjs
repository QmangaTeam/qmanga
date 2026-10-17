import {readFileSync} from 'node:fs';
const css=readFileSync('src/lib/shared/presentation/styles/tokens.css','utf8');
function luminance(hex) {
 const full=hex.length===4?'#'+[...hex.slice(1)].map(value=>value+value).join(''):hex;
 const rgb=[1,3,5].map(offset=>parseInt(full.slice(offset,offset+2),16)/255).map(value=>value<=.04045?value/12.92:((value+.055)/1.055)**2.4);
 return rgb[0]*.2126+rgb[1]*.7152+rgb[2]*.0722;
}
for(const [theme,selector] of [['dark',/:root \{([^}]+)\}/],['light',/:root\[data-theme='light'\] \{([^}]+)\}/]]) {
 const block=css.match(selector)?.[1];
 if(!block)throw new Error('Theme missing: '+theme);
 const tokens=Object.fromEntries([...block.matchAll(/--([a-z-]+):\s*(#[a-f\d]+)/gi)].map(match=>[match[1],match[2]]));
 for(const [foreground,background,minimum] of [['text','bg',4.5],['text','panel',4.5],['muted','bg',4.5],['muted','panel',4.5],['accent-ink','accent',4.5],['accent','bg',3],['muted','panel',3]]) {
  const values=[luminance(tokens[foreground]),luminance(tokens[background])].sort((left,right)=>right-left);
  const ratio=(values[0]+.05)/(values[1]+.05);
  if(ratio<minimum)throw new Error(theme+' '+foreground+'/'+background+' contrast '+ratio.toFixed(2)+' < '+minimum);
 }
}
console.log('Both theme palettes pass text, action and control contrast checks.');
