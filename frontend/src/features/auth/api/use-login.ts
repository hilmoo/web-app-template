import { useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AuthService } from "proto/servicesV1";
import { setMe } from "./me";

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation(AuthService.method.authLogin, {
    onSuccess: (res) => setMe(queryClient, res.data),
  });
}
