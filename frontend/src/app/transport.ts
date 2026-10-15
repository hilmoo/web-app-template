import { createConnectTransport } from "@connectrpc/connect-web";

// Same-origin: the core server (:8080) serves the API under /api and proxies Vite in dev,
// so the session cookie is sent automatically.
export const transport = createConnectTransport({ baseUrl: "/api" });
