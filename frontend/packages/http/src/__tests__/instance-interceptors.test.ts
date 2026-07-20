import { describe, it, expect, beforeEach, afterEach } from "vitest";
import {
  configureAuthTransport,
  resetAuthTransportConfiguration,
} from "../auth-transport-configuration";
import { instance } from "../instance.axios";
import { instanceWithoutInterceptors } from "../instance-without-interceptors.axios";
import AxiosMockAdapter from "axios-mock-adapter";

const mockInstance = new AxiosMockAdapter(instance);
const mockPublicInstance = new AxiosMockAdapter(instanceWithoutInterceptors);

describe("Instance interceptors", () => {
  beforeEach(() => {
    resetAuthTransportConfiguration();
    mockInstance.reset();
    mockPublicInstance.reset();
  });

  afterEach(() => {
    resetAuthTransportConfiguration();
  });

  it("private instance attaches Bearer token when configured", async () => {
    configureAuthTransport({
      getAccessToken: () => "my-bearer-token",
      getCsrfToken: () => null,
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    });

    mockInstance.onGet("/me").reply((config) => {
      const authHeader = config.headers?.["Authorization"];
      return [200, { header: authHeader }];
    });

    const response = await instance.get("/me");
    expect(response.data.header).toBe("Bearer my-bearer-token");
  });

  it("private instance does not attach Bearer when no config", async () => {
    mockInstance.onGet("/me").reply((config) => {
      const authHeader = config.headers?.["Authorization"];
      return [200, { header: authHeader }];
    });

    const response = await instance.get("/me");
    expect(response.data.header).toBeUndefined();
  });

  it("private instance attaches Bearer when token is null in config", async () => {
    configureAuthTransport({
      getAccessToken: () => null,
      getCsrfToken: () => null,
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    });

    mockInstance.onGet("/me").reply((config) => {
      const authHeader = config.headers?.["Authorization"];
      return [200, { header: authHeader }];
    });

    const response = await instance.get("/me");
    expect(response.data.header).toBeUndefined();
  });

  it("private instance does NOT attach CSRF token for GET", async () => {
    configureAuthTransport({
      getAccessToken: () => "token",
      getCsrfToken: () => "csrf-token",
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    });

    mockInstance.onGet("/me").reply((config) => {
      const csrfHeader = config.headers?.["X-CSRF-Token"];
      return [200, { csrf: csrfHeader }];
    });

    const response = await instance.get("/me");
    expect(response.data.csrf).toBeUndefined();
  });

  it("private instance attaches CSRF token for POST", async () => {
    configureAuthTransport({
      getAccessToken: () => "token",
      getCsrfToken: () => "csrf-token",
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    });

    mockInstance.onPost("/auth/logout").reply((config) => {
      const csrfHeader = config.headers?.["X-CSRF-Token"];
      return [200, { csrf: csrfHeader }];
    });

    const response = await instance.post("/auth/logout");
    expect(response.data.csrf).toBe("csrf-token");
  });

  it("public instance attaches CSRF for unsafe methods", async () => {
    configureAuthTransport({
      getAccessToken: () => null,
      getCsrfToken: () => "public-csrf",
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    });

    mockPublicInstance.onPost("/auth/login").reply((config) => {
      const csrfHeader = config.headers?.["X-CSRF-Token"];
      return [200, { csrf: csrfHeader }];
    });

    const response = await instanceWithoutInterceptors.post("/auth/login");
    expect(response.data.csrf).toBe("public-csrf");
  });

  it("public instance does NOT attach Bearer", async () => {
    configureAuthTransport({
      getAccessToken: () => "bearer-token",
      getCsrfToken: () => null,
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    });

    mockPublicInstance.onGet("/public").reply((config) => {
      const authHeader = config.headers?.["Authorization"];
      return [200, { header: authHeader }];
    });

    const response = await instanceWithoutInterceptors.get("/public");
    expect(response.data.header).toBeUndefined();
  });

  it("public instance does NOT attach CSRF for GET", async () => {
    configureAuthTransport({
      getAccessToken: () => null,
      getCsrfToken: () => "csrf-token",
      refresh: async () => {},
      shouldRefresh: () => false,
      shouldInvalidateAuth: () => false,
      onAuthLost: async () => {},
    });

    mockPublicInstance.onGet("/auth/csrf").reply((config) => {
      const csrfHeader = config.headers?.["X-CSRF-Token"];
      return [200, { csrf: csrfHeader }];
    });

    const response = await instanceWithoutInterceptors.get("/auth/csrf");
    expect(response.data.csrf).toBeUndefined();
  });
});
