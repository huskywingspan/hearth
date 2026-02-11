import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  type ReactNode,
} from 'react';
import pb from '@/lib/pocketbase';
import type { RecordModel } from 'pocketbase';
import { TOKEN_REFRESH_INTERVAL_MS } from '@/lib/constants';

interface AuthContextType {
  user: RecordModel | null;
  isValid: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, displayName: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<RecordModel | null>(pb.authStore.record);
  const [isValid, setIsValid] = useState(pb.authStore.isValid);
  const [isLoading, setIsLoading] = useState(true);

  // Listen for all auth state changes
  useEffect(() => {
    return pb.authStore.onChange((_token, record) => {
      setUser(record);
      setIsValid(pb.authStore.isValid);
    });
  }, []);

  // Validate existing token on mount
  useEffect(() => {
    async function validateToken() {
      if (pb.authStore.isValid) {
        try {
          await pb.collection('users').authRefresh();
        } catch {
          pb.authStore.clear();
        }
      }
      setIsLoading(false);
    }
    validateToken();
  }, []);

  // Auto-refresh token periodically
  useEffect(() => {
    const interval = setInterval(async () => {
      if (pb.authStore.isValid) {
        try {
          await pb.collection('users').authRefresh();
        } catch {
          pb.authStore.clear();
        }
      }
    }, TOKEN_REFRESH_INTERVAL_MS);
    return () => clearInterval(interval);
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    await pb.collection('users').authWithPassword(email, password);
  }, []);

  const register = useCallback(
    async (email: string, password: string, displayName: string) => {
      await pb.collection('users').create({
        email,
        password,
        passwordConfirm: password,
        display_name: displayName,
      });
      // Log in immediately after registration
      await pb.collection('users').authWithPassword(email, password);
    },
    []
  );

  const logout = useCallback(() => {
    pb.authStore.clear();
    // Unsubscribe all realtime subscriptions on logout
    pb.collection('messages').unsubscribe();
    pb.collection('rooms').unsubscribe();
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, isValid, isLoading, login, register, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
