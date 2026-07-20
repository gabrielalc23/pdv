import type { BroadcastAdapter } from "./coordinator.types";

export function createBroadcastAdapter(channelName: string): BroadcastAdapter {
  const channel =
    typeof BroadcastChannel !== "undefined"
      ? new BroadcastChannel(channelName)
      : null;

  return {
    postMessage(message: unknown): void {
      if (channel) {
        channel.postMessage(message);
      }
    },
    onMessage(handler: (message: unknown) => void): () => void {
      if (channel) {
        const wrapped = (event: MessageEvent): void => {
          handler(event.data);
        };
        channel.addEventListener("message", wrapped);
        return (): void => {
          channel.removeEventListener("message", wrapped);
        };
      }
      return (): void => {};
    },
    close(): void {
      if (channel) {
        channel.close();
      }
    },
  };
}
