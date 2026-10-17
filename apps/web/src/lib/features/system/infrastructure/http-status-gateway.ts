import type { StatusGateway, SystemStatus } from '../application/status';
export class HttpStatusGateway implements StatusGateway {
  constructor(private origin: string, private request: typeof fetch) {}
  async read(): Promise<SystemStatus> {
    const response = await this.request(this.origin + '/api/v1/health', { signal: AbortSignal.timeout(3000) });
    const result: unknown = await response.json();
    return { available: response.ok && typeof result === 'object' && result !== null && 'status' in result && result.status === 'ok' };
  }
}
