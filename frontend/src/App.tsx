import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Spinner } from '@/components/ui/Spinner';
import { AuthGuard } from '@/components/auth/AuthGuard';

// Code-split pages (K-024) — each page is a separate chunk
const ConnectPage = lazy(() => import('@/pages/ConnectPage'));
const LoginPage = lazy(() => import('@/pages/LoginPage'));
const HomePage = lazy(() => import('@/pages/HomePage'));
const RoomPage = lazy(() => import('@/pages/RoomPage'));
const DmPage = lazy(() => import('@/pages/DmPage'));

export default function App() {
  return (
    <BrowserRouter>
      <Suspense
        fallback={
          <div className="flex items-center justify-center h-screen">
            <Spinner />
          </div>
        }
      >
        <Routes>
          <Route path="/connect" element={<ConnectPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route element={<AuthGuard />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/room/:roomId" element={<RoomPage />} />
            <Route path="/dm/:dmId" element={<DmPage />} />
          </Route>
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
