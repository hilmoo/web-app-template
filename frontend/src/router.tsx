import { createRouter } from "@tanstack/react-router";
import { queryClient } from "./app/query-client";
import { transport } from "./app/transport";
import { routeTree } from "./routeTree.gen";

export const router = createRouter({
  routeTree,
  context: { queryClient, transport },
  defaultPreload: "intent",
  // React Query owns caching, so loaders should always run on preload/visit.
  defaultPreloadStaleTime: 0,
  scrollRestoration: true,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
