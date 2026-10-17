import { env } from '$env/dynamic/private';
import { HttpStatusGateway } from '$lib/features/system/infrastructure/http-status-gateway';
import { getStatus } from '$lib/features/system/application/status';
export function loadStatus(request: typeof fetch) {
  return getStatus(new HttpStatusGateway(env.API_INTERNAL_URL || 'http://api:8080', request));
}
