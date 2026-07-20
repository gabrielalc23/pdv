import { describe, it, expect, afterEach } from "vitest";
import { HttpMethod } from "../http-method.axios";
import { createApiCall } from "../create-api-call.axios";
import { instance } from "../instance.axios";
import { instanceWithoutInterceptors } from "../instance-without-interceptors.axios";
import { z } from "zod";
import { InvalidApiResponseError } from "@pdv/errors";
import AxiosMockAdapter from "axios-mock-adapter";

const mockInstance = new AxiosMockAdapter(instance);
const mockInstanceWithoutInterceptors = new AxiosMockAdapter(
  instanceWithoutInterceptors,
);

afterEach(() => {
  mockInstance.reset();
  mockInstanceWithoutInterceptors.reset();
});

describe("createApiCall", () => {
  it("sends GET request with params", async () => {
    mockInstanceWithoutInterceptors
      .onGet("/products", { params: { search: "foo" } })
      .reply(200, { data: ["filtered-foo"] });

    const api = createApiCall({
      type: "public",
      method: HttpMethod.GET,
      path: "/products",
      requestSchema: z.object({ search: z.string().optional() }),
      responseSchema: z.object({ data: z.array(z.string()) }),
    });

    const result = await api({ search: "foo" });

    expect(result.data).toEqual(["filtered-foo"]);
  });

  it("sends GET request without params", async () => {
    mockInstanceWithoutInterceptors
      .onGet("/products")
      .reply(200, { data: ["a", "b"] });

    const api = createApiCall({
      type: "public",
      method: HttpMethod.GET,
      path: "/products",
      requestSchema: z.object({}),
      responseSchema: z.object({ data: z.array(z.string()) }),
    });

    const result = await api({});

    expect(result.data).toEqual(["a", "b"]);
  });

  it("validates request payload against schema", async () => {
    const api = createApiCall({
      type: "public",
      method: HttpMethod.POST,
      path: "/test",
      requestSchema: z.object({ name: z.string().min(3) }),
      responseSchema: z.object({ id: z.string() }),
    });

    await expect(api({ name: "ab" })).rejects.toThrow();
  });

  it("throws InvalidApiResponseError when response doesn't match schema", async () => {
    mockInstanceWithoutInterceptors
      .onGet("/not-schema-match")
      .reply(200, { unexpected: "shape" });

    const api = createApiCall({
      type: "public",
      method: HttpMethod.GET,
      path: "/not-schema-match",
      requestSchema: z.object({}),
      responseSchema: z.object({ data: z.array(z.string()) }),
    });

    await expect(api({})).rejects.toThrow(InvalidApiResponseError);
  });
});

describe("http-method constants", () => {
  it("exports correct HTTP method constants", () => {
    expect(HttpMethod.GET).toBe("GET");
    expect(HttpMethod.POST).toBe("POST");
    expect(HttpMethod.PUT).toBe("PUT");
    expect(HttpMethod.PATCH).toBe("PATCH");
    expect(HttpMethod.DELETE).toBe("DELETE");
  });
});

describe("requestLocation inference", () => {
  it("infers 'params' from GET", async () => {
    mockInstanceWithoutInterceptors
      .onGet("/search", { params: { q: "test" } })
      .reply(200, { results: ["a", "b"] });

    const api = createApiCall({
      type: "public",
      method: HttpMethod.GET,
      path: "/search",
      requestSchema: z.object({ q: z.string() }),
      responseSchema: z.object({ results: z.array(z.string()) }),
    });

    const result = await api({ q: "test" });

    expect(result.results).toEqual(["a", "b"]);
  });

  it("infers 'data' from POST", async () => {
    mockInstanceWithoutInterceptors
      .onPost("/test")
      .reply(200, { created: true });

    const api = createApiCall({
      type: "public",
      method: HttpMethod.POST,
      path: "/test",
      requestSchema: z.object({ value: z.number() }),
      responseSchema: z.object({ created: z.boolean() }),
    });

    const result = await api({ value: 42 });

    expect(result.created).toBe(true);
  });
});
