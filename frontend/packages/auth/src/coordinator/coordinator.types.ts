export interface LockAdapter {
  acquire: (name: string, callback: () => Promise<void>) => Promise<void>;
  isSupported: boolean;
}

export interface BroadcastAdapter {
  postMessage: (message: unknown) => void;
  onMessage: (handler: (message: unknown) => void) => () => void;
  close: VoidFunction;
}
