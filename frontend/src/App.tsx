import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { ApolloProvider, type ApolloClient, type NormalizedCacheObject } from '@apollo/client';
import { client as defaultClient } from './utils/apollo-client';
import { AuthProvider } from './contexts/AuthContext';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { Login } from './components/auth/Login';
import { Register } from './components/auth/Register';
import { Layout } from './components/layout/Layout';
import { ErrorBoundary } from './components/common/ErrorBoundary';
import { Dashboard } from './pages/Dashboard';
import { Team } from './pages/Team';
import { Profile } from './pages/Profile';

export const AppRoutes = () => (
  <Routes>
    <Route path="/login" element={<Login />} />
    <Route path="/register" element={<Register />} />
    <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
      <Route path="/dashboard" element={<Dashboard />} />
      <Route path="/team" element={<Team />} />
      <Route path="/profile" element={<Profile />} />
    </Route>
    <Route path="*" element={<Navigate to="/dashboard" replace />} />
  </Routes>
);

function App({ client = defaultClient }: { client?: ApolloClient<NormalizedCacheObject> }) {
  return (
    <ErrorBoundary>
      <ApolloProvider client={client}>
        <AuthProvider>
          <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
            <AppRoutes />
          </BrowserRouter>
        </AuthProvider>
      </ApolloProvider>
    </ErrorBoundary>
  );
}

export default App;
