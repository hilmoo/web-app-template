import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AuthService } from "proto/servicesCoreV1";

export function useAuthLogout() {
  const queryClient = useQueryClient();
  return useMutation(AuthService.method.authLogout, {
    onSuccess: () => queryClient.clear(),
  });
}
