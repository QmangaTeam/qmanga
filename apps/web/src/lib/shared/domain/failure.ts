export class AppFailure extends Error {
 constructor(readonly code:string) { super(code); this.name='AppFailure'; }
}
