import { describe, it, expect, beforeEach, afterEach } from "vitest";
import {
  configureAuthTransport,
  resetAuthTransportConfiguration,
  getAuthTransportConfiguration,
} from "../auth-transport-configuration";
import type { AuthTransportConfiguration } from "../auth-transport-configuration";

describe("AuthTransportConfiguration", () => {
  beforeEach(() => {
    resetAuthTransportConfiguration();
  });

  afterEach(() => {
    resetAuthTransportConfiguration();
  });

  it("starts with null config", () => {
    expect(getAuthTransportConfiguration()).toBeNull();
  });

  it("sets config via configureAuthTransport", () => {
    const config: AuthTransportConfiguration = {
      getAccessToken: () => "test-token",
      getCsrfToken: () => null,
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    };

    configureAuthTransport(config);
    expect(getAuthTransportConfiguration()).toBe(config);
  });

  it("cleanup removes config", () => {
    const config: AuthTransportConfiguration = {
      getAccessToken: () => "test-token",
      getCsrfToken: () => null,
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    };

    const cleanup = configureAuthTransport(config);
    expect(getAuthTransportConfiguration()).toBe(config);

    cleanup();
    expect(getAuthTransportConfiguration()).toBeNull();
  });

  it("cleanup does not affect other config", () => {
    const config1: AuthTransportConfiguration = {
      getAccessToken: () => "token1",
      getCsrfToken: () => null,
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    };
    const config2: AuthTransportConfiguration = {
      getAccessToken: () => "token2",
      getCsrfToken: () => null,
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    };

    const cleanup1 = configureAuthTransport(config1);
    configureAuthTransport(config2);
    expect(getAuthTransportConfiguration()).toBe(config2);

    cleanup1(); // should not remove config2
    expect(getAuthTransportConfiguration()).toBe(config2);
  });

  it("is not imported from @pdv/auth", () => {
    // Structural test: the transport config module should not import from @pdv/auth
    const fs = require("node:fs");
    const source = fs.readFileSync(
      __dirname + "/../auth-transport-configuration.ts",
      "utf-8",
    );
    expect(source).not.toContain("@pdv/auth");
  });
});
