import { useState, type FormEvent } from 'react';
import { useMutation } from '@apollo/client';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { REGISTER } from '../../graphql/mutations';
import { useAuth } from '../../contexts/AuthContext';
import { getErrorMessage } from '../../utils/errors';
import type { User } from '../../types';
import { validateRegistration, type RegistrationForm as FormState } from '../../utils/validation';

const inputClass =
  'block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm';

export const Register = () => {
  const navigate = useNavigate();
  const { login, isAuthenticated } = useAuth();
  const [form, setForm] = useState<FormState>({ name: '', email: '', password: '', confirmPassword: '' });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [error, setError] = useState('');
  const [mutate, { loading }] = useMutation<{ register: { token: string; refreshToken: string; user: User } }>(REGISTER);

  if (isAuthenticated) return <Navigate to="/dashboard" replace />;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    const errors = validateRegistration(form);
    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) return;
    try {
      const { data } = await mutate({
        variables: { input: { name: form.name.trim(), email: form.email.trim(), password: form.password } },
      });
      if (data) {
        login(data.register.token, data.register.refreshToken, data.register.user);
        navigate('/dashboard');
      }
    } catch (err) {
      setError(getErrorMessage(err, 'Registration failed. Please try again.'));
    }
  };

  const field = (name: keyof FormState, label: string, type = 'text', autoComplete?: string) => (
    <div>
      <label htmlFor={name} className="mb-1 block text-sm font-medium text-gray-700">
        {label}
      </label>
      <input id={name} type={type} autoComplete={autoComplete} className={inputClass} value={form[name]} disabled={loading}
        onChange={(e) => setForm({ ...form, [name]: e.target.value })} />
      {fieldErrors[name] && <p className="mt-1 text-sm text-red-600">{fieldErrors[name]}</p>}
    </div>
  );

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4 py-12">
      <div className="w-full max-w-md space-y-8 rounded-xl bg-white p-10 shadow-lg">
        <div>
          <h2 className="text-center text-3xl font-extrabold text-gray-900">Create your account</h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Already have an account?{' '}
            <Link to="/login" className="font-medium text-indigo-600 hover:text-indigo-500">
              Sign in
            </Link>
          </p>
        </div>
        <form className="space-y-4" onSubmit={handleSubmit} noValidate>
          {error && (
            <div role="alert" className="rounded-md bg-red-50 p-3 text-sm text-red-800">
              {error}
            </div>
          )}
          {field('name', 'Full name', 'text', 'name')}
          {field('email', 'Email address', 'email', 'email')}
          {field('password', 'Password', 'password', 'new-password')}
          {field('confirmPassword', 'Confirm password', 'password', 'new-password')}
          <p className="text-xs text-gray-500">
            At least 8 characters with upper- and lower-case letters, a number and a special character.
          </p>
          <button type="submit" disabled={loading}
            className="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50">
            {loading ? 'Creating account...' : 'Create account'}
          </button>
        </form>
      </div>
    </div>
  );
};
