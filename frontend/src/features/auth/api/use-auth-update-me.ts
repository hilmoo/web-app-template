import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AuthService } from "proto/servicesV1";
import { setAuthMe } from "./use-auth-me";

export function useAuthUpdateMe() {
  const queryClient = useQueryClient();
  return useMutation(AuthService.method.authUpdateMe, {
    onSuccess: (res) => setAuthMe(queryClient, res.data),
  });
}
