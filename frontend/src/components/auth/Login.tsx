import { useState, type FormEvent } from 'react';
import { useMutation } from '@apollo/client';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { LOGIN } from '../../graphql/mutations';
import { useAuth } from '../../contexts/AuthContext';
import { getErrorMessage } from '../../utils/errors';
import type { User } from '../../types';

const inputClass =
  'block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm';

export const Login = () => {
  const navigate = useNavigate();
  const { login, isAuthenticated } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const [mutate, { loading }] = useMutation<{ login: { token: string; refreshToken: string; user: User } }>(LOGIN);

  if (isAuthenticated) return <Navigate to="/dashboard" replace />;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    if (!email || !password) {
      setError('Please fill in all fields');
      return;
    }
    try {
      const { data } = await mutate({ variables: { input: { email, password } } });
      if (data) {
        login(data.login.token, data.login.refreshToken, data.login.user);
        navigate('/dashboard');
      }
    } catch (err) {
      setError(getErrorMessage(err, 'Login failed. Please check your credentials.'));
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4 py-12">
      <div className="w-full max-w-md space-y-8 rounded-xl bg-white p-10 shadow-lg">
        <div>
          <h2 className="text-center text-3xl font-extrabold text-gray-900">Sign in to TaskBoard</h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Or{' '}
            <Link to="/register" className="font-medium text-indigo-600 hover:text-indigo-500">
              create a new account
            </Link>
          </p>
        </div>
        <form className="space-y-4" onSubmit={handleSubmit} noValidate>
          {error && (
            <div role="alert" className="rounded-md bg-red-50 p-3 text-sm text-red-800">
              {error}
            </div>
          )}
          <div>
            <label htmlFor="email" className="mb-1 block text-sm font-medium text-gray-700">
              Email address
            </label>
            <input id="email" type="email" autoComplete="email" className={inputClass} value={email}
              onChange={(e) => setEmail(e.target.value)} disabled={loading} />
          </div>
          <div>
            <label htmlFor="password" className="mb-1 block text-sm font-medium text-gray-700">
              Password
            </label>
            <input id="password" type="password" autoComplete="current-password" className={inputClass} value={password}
              onChange={(e) => setPassword(e.target.value)} disabled={loading} />
          </div>
          <button type="submit" disabled={loading}
            className="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50">
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>
      </div>
    </div>
  );
};
