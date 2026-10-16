import { lazy, Suspense } from "react";
import type { QueryClient } from "@tanstack/react-query";
import type { Transport } from "@connectrpc/connect";
import { Outlet, createRootRouteWithContext } from "@tanstack/react-router";

// Devtools are dev-only: the import.meta.env.DEV guard lets the bundler drop them from production builds.
const Devtools = import.meta.env.DEV
  ? lazy(() => import("#/shared/components/devtools"))
  : () => null;

export interface RouterContext {
  queryClient: QueryClient;
  transport: Transport;
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootComponent,
});

function RootComponent() {
  return (
    <>
      <Outlet />
      <Suspense fallback={null}>
        <Devtools />
      </Suspense>
    </>
  );
}
