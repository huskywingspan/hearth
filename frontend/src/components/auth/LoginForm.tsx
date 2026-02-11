import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';
import { useAuth } from '@/hooks/useAuth';

/**
 * Login/register form — warm, inviting, pillow buttons.
 * Switches between login and register modes.
 */
export function LoginForm() {
  const { login, register } = useAuth();
  const navigate = useNavigate();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      if (mode === 'login') {
        await login(email, password);
      } else {
        await register(email, password, displayName || email.split('@')[0]!);
      }
      navigate('/', { replace: true });
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'Something went wrong. Please try again.'
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card className="w-full max-w-sm mx-auto animate-float-in">
      <h1 className="font-display text-2xl text-center mb-6 text-[var(--color-text-primary)]">
        {mode === 'login' ? 'Welcome Home' : 'Find Your Hearth'}
      </h1>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        {mode === 'register' && (
          <Input
            label="Display Name"
            placeholder="Your name by the fire"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
          />
        )}

        <Input
          label="Email"
          type="email"
          placeholder="you@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />

        <Input
          label="Password"
          type="password"
          placeholder="••••••••"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          error={error || undefined}
        />

        <Button
          type="submit"
          variant="primary"
          size="lg"
          className="mt-2 w-full"
          disabled={isSubmitting}
        >
          {isSubmitting
            ? '...'
            : mode === 'login'
              ? 'Enter'
              : 'Join'}
        </Button>
      </form>

      <div className="mt-6 text-center">
        <button
          type="button"
          onClick={() => {
            setMode(mode === 'login' ? 'register' : 'login');
            setError('');
          }}
          className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-accent-amber)] transition-colors duration-[var(--duration-normal)]"
        >
          {mode === 'login'
            ? "Don't have a hearth? Create one"
            : 'Already have a hearth? Welcome back'}
        </button>
      </div>
    </Card>
  );
}
