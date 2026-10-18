export interface Account { displayName?:string; id: string; email: string; role: 'reader' | 'moderator' | 'admin'; verified: boolean }
export interface Credentials { email: string; password: string; displayName?: string }
export function credentialError(input: Credentials, registration: boolean): 'email' | 'password' | 'nickname' | null {
 if (!/^[^@\s]+@[^@\s]+$/.test(input.email.trim()) || input.email.length > 254) return 'email';
 if (registration && !validNickname(input.displayName || '')) return 'nickname';
 const length = [...input.password].length;
 if (!length || input.password.length > 512 || registration && (length < 10 || length > 128)) return 'password';
 return null;
}

export function validNickname(value:string):boolean { return [...value].length>=2 && [...value].length<=40 && value===value.trim() && !value.includes('  ') && /^[\p{L}\p{N} _.-]+$/u.test(value); }
