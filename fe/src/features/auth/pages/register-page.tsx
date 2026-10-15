import { useNavigate } from "@tanstack/react-router";
import { AuthLayout } from "../components/auth-layout";
import { RegisterForm } from "../components/register-form";

export function RegisterPage() {
  const navigate = useNavigate();
  return (
    <AuthLayout
      title="Create your account"
      footerText="Already have an account?"
      footerLink={{ to: "/login", label: "Sign in" }}
    >
      <RegisterForm onSuccess={() => navigate({ to: "/" })} />
    </AuthLayout>
  );
}
