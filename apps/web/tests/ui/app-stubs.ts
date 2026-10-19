export async function goto(): Promise<void> {}
export async function invalidateAll(): Promise<void> {}
export function beforeNavigate(): void {}
export function afterNavigate(): void {}
export const page = { url: new URL('http://localhost/') };
