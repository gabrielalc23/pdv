import { createAuthRuntime, type AuthRuntime } from "@pdv/auth";
import { instance } from "@pdv/http";

let runtimeInstance: AuthRuntime | null = null;

export function getAuthRuntime(): AuthRuntime {
  if (!runtimeInstance) {
    runtimeInstance = createAuthRuntime({
      publicInstance: instance,
      channelName: "pdv-auth-admin",
      onAuthLost: (_reason) => {
        // Will be connected in Task 15
      },
    });
  }
  return runtimeInstance;
}

export function destroyAuthRuntime(): void {
  if (runtimeInstance) {
    runtimeInstance.destroy();
    runtimeInstance = null;
  }
}
