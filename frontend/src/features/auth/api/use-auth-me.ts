import { create } from "@bufbuild/protobuf";
import { ConnectError, Code, type Transport } from "@connectrpc/connect";
import { createConnectQueryKey, createQueryOptions, useQuery } from "@connectrpc/connect-query";
import type { QueryClient } from "@tanstack/react-query";
import { AuthMeResponseSchema, AuthService } from "proto/servicesV1";
import type { AuthMeUser } from "proto/modelsV1";

const authMe = AuthService.method.authMe;

export const authMeQueryKey = () =>
  createConnectQueryKey({ schema: authMe, input: {}, cardinality: "finite" });

export function setAuthMe(queryClient: QueryClient, user: AuthMeUser | undefined) {
  queryClient.setQueryData(authMeQueryKey(), create(AuthMeResponseSchema, { data: user }));
}

export function useAuthMe() {
  return useQuery(authMe, {}, { select: (res) => res.data });
}

export async function fetchAuthMe(queryClient: QueryClient, transport: Transport) {
  try {
    const res = await queryClient.query({
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
