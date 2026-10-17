import { env } from '$env/dynamic/private';
import { HttpStatusGateway } from '$lib/features/system/infrastructure/http-status-gateway';
import { getStatus } from '$lib/features/system/application/status';
import { HttpAuthGateway } from '$lib/features/identity/infrastructure/http-auth-gateway';
import { HttpCatalogGateway } from '$lib/features/catalog/infrastructure/http-catalog-gateway';
import { ApiClient } from '$lib/shared/infrastructure/api-client';
import type { CatalogGateway } from '$lib/features/catalog/application/catalog';
export function createPublicCatalogue(request: typeof fetch): CatalogGateway {
 const base = env.API_INTERNAL_URL || 'http://api:8080';
 return new HttpCatalogGateway(new ApiClient((input, init) => request(base + String(input), init)));
}
export function loadAccount(request: typeof fetch, cookie: string) {
  if (!cookie.includes('qmanga_session=')) return Promise.resolve(null);
  return new HttpAuthGateway(env.API_INTERNAL_URL || 'http://api:8080', request, cookie).current();
}
export function loadStatus(request: typeof fetch) {
  return getStatus(new HttpStatusGateway(env.API_INTERNAL_URL || 'http://api:8080', request));
}
