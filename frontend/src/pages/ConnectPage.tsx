import { useState } from 'react';
import { setServerUrl } from '@/lib/pocketbase';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';

/**
 * ConnectPage — server URL entry screen.
 * Shown when no server URL is configured and the app isn't served from PocketBase.
 * The user enters their House address (e.g., https://abc123.trycloudflare.com).
 */
export default function ConnectPage() {
  const [url, setUrl] = useState('');
  const [error, setError] = useState('');
  const [checking, setChecking] = useState(false);

  const handleConnect = async () => {
    if (!url.trim()) return;
    setChecking(true);
    setError('');

    try {
      // Normalize URL
      let target = url.trim().replace(/\/+$/, '');
      if (!/^https?:\/\//i.test(target)) target = 'https://' + target;

      // Ping PocketBase health endpoint to verify it's reachable
      const res = await fetch(`${target}/api/health`, {
        method: 'GET',
        signal: AbortSignal.timeout(5000),
      });

      if (!res.ok) throw new Error('Not a Hearth server');

      setServerUrl(target); // Stores in localStorage and reloads
    } catch {
      setError('Could not reach that server. Check the URL and try again.');
      setChecking(false);
    }
  };

  return (
    <div className="flex items-center justify-center h-screen bg-[var(--color-bg-primary)] p-4">
      <Card className="max-w-md w-full animate-float-in">
        <div className="text-center mb-6">
          <h1 className="font-display text-3xl text-[var(--color-accent-amber)] mb-2">
            Hearth
          </h1>
          <p className="text-[var(--color-text-secondary)]">
            Enter your House address to connect.
          </p>
        </div>

        <div className="space-y-4">
          <Input
            type="url"
            placeholder="https://hearth.example.com"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleConnect()}
            disabled={checking}
            autoFocus
          />

          {error && (
            <p className="text-sm text-[var(--color-alert-clay)]">{error}</p>
          )}

          <Button onClick={handleConnect} disabled={checking} className="w-full">
            {checking ? 'Connecting...' : 'Enter House'}
          </Button>
        </div>

        <p className="text-xs text-[var(--color-text-muted)] mt-4 text-center">
          Ask your Homeowner for the House URL, or scan a QR code.
        </p>
      </Card>
    </div>
  );
}
