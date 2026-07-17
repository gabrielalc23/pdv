export type LogPayload = Record<string, unknown>;

export function sendLog(data: LogPayload): void {
  const blob: Blob = new Blob([JSON.stringify(data)], {
    type: "application/json",
  });
  navigator.sendBeacon("/api/log-endpoint", blob);
}
