import { LoginForm } from '@/components/auth/LoginForm';

export default function LoginPage() {
  return (
    <div className="flex items-center justify-center min-h-screen p-4 bg-[var(--color-bg-primary)]">
      <LoginForm />
    </div>
  );
}
