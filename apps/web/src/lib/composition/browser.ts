import { HttpAuthGateway } from '$lib/features/identity/infrastructure/http-auth-gateway';
export function createAuth() { return new HttpAuthGateway('', globalThis.fetch.bind(globalThis)); }
