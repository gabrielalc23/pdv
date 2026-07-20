import type { LockAdapter } from "./coordinator.types";

export function createLockAdapter(): LockAdapter {
  const hasLocks =
    typeof navigator !== "undefined" &&
    typeof (
      navigator as {
        locks?: {
          request: (
            name: string,
            callback: () => Promise<void>,
          ) => Promise<void>;
        };
      }
    ).locks?.request === "function";

  if (hasLocks) {
    return {
      acquire: async (
        name: string,
        callback: () => Promise<void>,
      ): Promise<void> => {
        await (
          navigator as {
            locks: {
              request: (
                name: string,
                callback: () => Promise<void>,
              ) => Promise<void>;
            };
          }
        ).locks.request(name, callback);
      },
      isSupported: true,
    };
  }

  return {
    acquire: async (
      _name: string,
      callback: () => Promise<void>,
    ): Promise<void> => {
      await callback();
    },
    isSupported: false,
  };
}
