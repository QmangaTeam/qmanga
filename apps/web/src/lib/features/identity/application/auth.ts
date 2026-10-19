import type { Account, Credentials } from '../domain/account';
import { credentialError } from '../domain/account';
export interface AuthResult { ok: boolean; message: string; code?: string; account?: Account }
export interface AuthGateway {
 completeExternalLogin?():Promise<void>;
 googleStatus?():Promise<boolean>;
 googleStart?(returnPath:string,nickname:string):Promise<{url?:string;code?:string}>;
 updateProfile(accountId:string,displayName:string):Promise<AuthResult>;
 login(input: Credentials): Promise<AuthResult>;
 register(input: Credentials): Promise<AuthResult>;
 current(): Promise<Account | null>;
 requestReset(email: string): Promise<AuthResult>;
 resend(email: string): Promise<AuthResult>;
 verify(token: string): Promise<AuthResult>;
 reset(token: string, password: string): Promise<AuthResult>;
 logout(): Promise<AuthResult>;
}
export async function authenticate(gateway: AuthGateway, mode: 'login' | 'register', input: Credentials): Promise<AuthResult> {
 const invalid = credentialError(input, mode === 'register');
 if (invalid) return { ok: false, code: invalid, message: invalid === 'email' ? 'Пошта мекенжайын тексеріңіз' : 'Құпиясөзді тексеріңіз. 8–128 таңба қолданыңыз.' };
 return gateway[mode]({ email: input.email.trim().toLowerCase(), password: input.password, displayName:input.displayName });
}
