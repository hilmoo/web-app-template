import { create } from "@bufbuild/protobuf";
import { ConnectError, Code, type Transport } from "@connectrpc/connect";
import { createConnectQueryKey, createQueryOptions, useQuery } from "@connectrpc/connect-query";
import type { QueryClient } from "@tanstack/react-query";
import { AuthMeResponseSchema, AuthService } from "proto/servicesV1";
import type { AuthMeUser } from "proto/modelsV1";

const authMe = AuthService.method.authMe;

export const meQueryKey = () =>
  createConnectQueryKey({ schema: authMe, input: {}, cardinality: "finite" });

export function setMe(queryClient: QueryClient, user: AuthMeUser | undefined) {
  queryClient.setQueryData(meQueryKey(), create(AuthMeResponseSchema, { data: user }));
}

export function useMe() {
  return useQuery(authMe, {}, { select: (res) => res.data });
}

export async function fetchMe(queryClient: QueryClient, transport: Transport) {
  try {
    const res = await queryClient.fetchQuery({
      ...createQueryOptions(authMe, {}, { transport }),
      staleTime: 30_000,
    });
    return res.data ?? null;
  } catch (error) {
    const code = ConnectError.from(error).code;
    if (code === Code.Unauthenticated || code === Code.PermissionDenied) {
      return null;
    }
    throw error;
  }
}
