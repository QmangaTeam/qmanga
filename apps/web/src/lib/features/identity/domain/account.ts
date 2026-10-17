export interface Account { id: string; email: string; role: 'reader' | 'moderator' | 'admin'; verified: boolean }
export interface Credentials { email: string; password: string }
export function credentialError(input: Credentials, registration: boolean): 'email' | 'password' | null {
 if (!/^[^@\s]+@[^@\s]+$/.test(input.email.trim()) || input.email.length > 254) return 'email';
 const length = [...input.password].length;
 if (!length || input.password.length > 512 || registration && (length < 10 || length > 128)) return 'password';
 return null;
}
