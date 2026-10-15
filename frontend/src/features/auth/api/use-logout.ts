import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AuthService } from "proto/servicesV1";

export function useLogout() {
  const queryClient = useQueryClient();
  return useMutation(AuthService.method.authLogout, {
    onSuccess: () => queryClient.clear(),
  });
}
