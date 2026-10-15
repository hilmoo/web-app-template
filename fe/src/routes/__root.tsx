import type { QueryClient } from "@tanstack/react-query";
import type { Transport } from "@connectrpc/connect";
import { Outlet, createRootRouteWithContext } from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { TanStackDevtools } from "@tanstack/react-devtools";

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
      <TanStackDevtools
        config={{
          position: "bottom-right",
        }}
        plugins={[
          {
            name: "TanStack Router",
            render: <TanStackRouterDevtoolsPanel />,
          },
          {
            name: "TanStack Query",
            render: <ReactQueryDevtools />,
          },
        ]}
      />
    </>
  );
}
