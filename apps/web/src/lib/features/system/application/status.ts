export interface SystemStatus { available: boolean }
export interface StatusGateway { read(): Promise<SystemStatus> }
export async function getStatus(gateway: StatusGateway): Promise<SystemStatus> {
  try { return await gateway.read(); } catch { return { available: false }; }
}
