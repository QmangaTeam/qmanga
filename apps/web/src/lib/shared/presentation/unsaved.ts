/** Confirm only a transition that would discard actual local edits. */
export function allowDiscard(dirty:boolean,ask:()=>boolean):boolean{return !dirty||ask();}
