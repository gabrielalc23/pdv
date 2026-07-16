import type { ReactNode } from "react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render as rtlRender } from "@testing-library/react"
import { http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import type { SetupServer } from "msw/node"
import type { DefaultBodyType } from "msw"

const API_BASE_URL = "http://localhost:3000"

export function createMockHandler<TBody extends DefaultBodyType>(
  method: "get" | "post" | "put" | "delete" | "patch",
  path: string,
  status: number,
  body: TBody,
) {
  const url = `${API_BASE_URL}${path}`
  return http[method](url, () => HttpResponse.json(body, { status: status as any }))
}

export function createTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  })
}

interface TestWrapperProps {
  children: ReactNode
  client?: QueryClient
}

export function TestWrapper({ children, client }: TestWrapperProps) {
  const queryClient: QueryClient = client ?? createTestQueryClient()

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}

export function renderWithProviders(ui: ReactNode, client?: QueryClient) {
  const queryClient: QueryClient = client ?? createTestQueryClient()

  return rtlRender(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>,
  )
}

export const testServer: SetupServer = setupServer()
