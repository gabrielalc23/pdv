import { setupServer } from "msw/node";
import { http, HttpResponse } from "msw";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000";

export const handlers = [
  http.get(`${API_BASE_URL}/products`, ({ request }) => {
    const url = new URL(request.url);
    const search = url.searchParams.get("search");
    return HttpResponse.json({
      data: search ? [`filtered-${search}`] : ["a", "b"],
    });
  }),
];

export const server = setupServer(...handlers);

server.listen();

self.addEventListener("beforeunload", () => {
  server.close();
});
