import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AuthService } from "proto/servicesCoreV1";
import { setAuthMe } from "./use-auth-me";

export function useAuthRegister() {
  const queryClient = useQueryClient();
  return useMutation(AuthService.method.authRegister, {
    onSuccess: (res) => setAuthMe(queryClient, res.data),
  });
}
