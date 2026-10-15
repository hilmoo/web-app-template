import { useNavigate } from "@tanstack/react-router";
import { AuthLayout } from "../components/auth-layout";
import { LoginForm } from "../components/login-form";

export function LoginPage() {
  const navigate = useNavigate();
  return (
    <AuthLayout
      title="Welcome back"
      footerText="Don't have an account?"
      footerLink={{ to: "/register", label: "Create account" }}
    >
      <LoginForm onSuccess={() => navigate({ to: "/" })} />
    </AuthLayout>
  );
}
