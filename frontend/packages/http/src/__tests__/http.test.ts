import { describe, it, expect } from "vitest"
import { HttpMethod } from "../http-method.axios"
import { createApiCall } from "../create-api-call.axios"
import { z } from "zod"
import { InvalidApiResponseError } from "@pdv/errors"
import { setupServer } from "msw/native"
import { http, HttpResponse } from "msw"

describe("createApiCall", () => {
  it("sends GET request with params", async () => {
    const api = createApiCall({
      type: "public",
      method: HttpMethod.GET,
      path: "/products",
      requestSchema: z.object({ search: z.string().optional() }),
      responseSchema: z.object({ data: z.array(z.string()) }),
    })

    const result = await api({ search: "foo" })

    expect(result.data).toEqual(["filtered-foo"])
  })

  it("sends GET request without params", async () => {
    const api = createApiCall({
      type: "public",
      method: HttpMethod.GET,
      path: "/products",
      requestSchema: z.object({}),
      responseSchema: z.object({ data: z.array(z.string()) }),
    })

    const result = await api({})

    expect(result.data).toEqual(["a", "b"])
  })

  it("validates request payload against schema", async () => {
    const api = createApiCall({
      type: "public",
      method: HttpMethod.POST,
      path: "/test",
      requestSchema: z.object({ name: z.string().min(3) }),
      responseSchema: z.object({ id: z.string() }),
    })

    await expect(api({ name: "ab" })).rejects.toThrow()
  })

  it("throws InvalidApiResponseError when response doesn't match schema", async () => {
    
    const mockServer = setupServer(
      http.get("/not-schema-match", () => HttpResponse.json({ unexpected: "shape" }))
    )
    mockServer.listen()

    const api = createApiCall({
      type: "public",
      method: HttpMethod.GET,
      path: "/not-schema-match",
      requestSchema: z.object({}),
      responseSchema: z.object({ data: z.array(z.string()) }),
    })

    await expect(api({})).rejects.toThrow(InvalidApiResponseError)

    mockServer.close()
  })
})

describe("http-method constants", () => {
  it("exports correct HTTP method constants", () => {
    expect(HttpMethod.GET).toBe("GET")
    expect(HttpMethod.POST).toBe("POST")
    expect(HttpMethod.PUT).toBe("PUT")
    expect(HttpMethod.PATCH).toBe("PATCH")
    expect(HttpMethod.DELETE).toBe("DELETE")
  })
})

describe("requestLocation inference", () => {
  it("infers 'params' from GET", async () => {
    const api = createApiCall({
      type: "public",
      method: HttpMethod.GET,
      path: "/search",
      requestSchema: z.object({ q: z.string() }),
      responseSchema: z.object({ results: z.array(z.string()) }),
    })

    const result = await api({ q: "test" })

    expect(result.results).toEqual(["a", "b"])
  })

  it("infers 'data' from POST", async () => {
    const api = createApiCall({
      type: "public",
      method: HttpMethod.POST,
      path: "/test",
      requestSchema: z.object({ value: z.number() }),
      responseSchema: z.object({ created: z.boolean() }),
    })

    const result = await api({ value: 42 })

    expect(result.created).toBe(true)
  })
})