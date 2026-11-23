import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MockedProvider } from '@apollo/client/testing';
import { GraphQLError } from 'graphql';
import { Route } from 'react-router-dom';
import { Login } from '../src/components/auth/Login';
import { Register } from '../src/components/auth/Register';
import { ProtectedRoute } from '../src/components/auth/ProtectedRoute';
import { AuthProvider, REFRESH_INTERVAL_MS, useAuth } from '../src/contexts/AuthContext';
import { useAuth as useAuthHook } from '../src/hooks/useAuth';
import { LOGIN, REGISTER } from '../src/graphql/mutations';
import { LOGOUT_EVENT } from '../src/utils/apollo-client';
import * as apollo from '../src/utils/apollo-client';
import { alice, signIn } from './fixtures';
import { renderWithProviders } from './render';

const dashboard = <Route path="/dashboard" element={<p>Dashboard page</p>} />;
const PW = 'Str0ng!Pass';

describe('Login', () => {
  const mock = (result: object) => ({
    request: { query: LOGIN, variables: { input: { email: 'alice@example.com', password: PW } } },
    ...result,
  });

  it('signs in and goes to the dashboard', async () => {
    const user = userEvent.setup();
    renderWithProviders(<Login />, {
      route: '/login', path: '/login', extraRoutes: dashboard,
      mocks: [mock({ result: { data: { login: { __typename: 'AuthPayload', token: 't', refreshToken: 'r', user: alice } } } })],
    });
    await user.type(screen.getByLabelText('Email address'), 'alice@example.com');
    await user.type(screen.getByLabelText('Password'), PW);
    await user.click(screen.getByRole('button', { name: 'Sign in' }));
    expect(await screen.findByText('Dashboard page')).toBeInTheDocument();
    expect(localStorage.getItem('accessToken')).toBe('t');
  });

  it('requires both fields', async () => {
    const user = userEvent.setup();
    renderWithProviders(<Login />, { route: '/login', path: '/login' });
    await user.click(screen.getByRole('button', { name: 'Sign in' }));
    expect(screen.getByRole('alert')).toHaveTextContent('Please fill in all fields');
  });

  it('shows the server error', async () => {
    const user = userEvent.setup();
    renderWithProviders(<Login />, {
      route: '/login', path: '/login',
      mocks: [mock({ result: { errors: [new GraphQLError('invalid email or password')] } })],
    });
    await user.type(screen.getByLabelText('Email address'), 'alice@example.com');
    await user.type(screen.getByLabelText('Password'), PW);
    await user.click(screen.getByRole('button', { name: 'Sign in' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('invalid email or password');
  });

  it('redirects when already signed in', async () => {
    signIn();
    renderWithProviders(<Login />, { route: '/login', path: '/login', extraRoutes: dashboard });
    expect(await screen.findByText('Dashboard page')).toBeInTheDocument();
  });
});

describe('Register', () => {
  const input = { name: 'Alice', email: 'alice@example.com', password: PW };
  const fill = async (user: ReturnType<typeof userEvent.setup>, confirm = PW) => {
    await user.type(screen.getByLabelText('Full name'), 'Alice');
    await user.type(screen.getByLabelText('Email address'), 'alice@example.com');
    await user.type(screen.getByLabelText('Password'), PW);
    await user.type(screen.getByLabelText('Confirm password'), confirm);
  };

  it('creates an account and signs in', async () => {
    const user = userEvent.setup();
    renderWithProviders(<Register />, {
      route: '/register', path: '/register', extraRoutes: dashboard,
      mocks: [{ request: { query: REGISTER, variables: { input } },
        result: { data: { register: { __typename: 'AuthPayload', token: 't', refreshToken: 'r', user: alice } } } }],
    });
    await fill(user);
    await user.click(screen.getByRole('button', { name: 'Create account' }));
    expect(await screen.findByText('Dashboard page')).toBeInTheDocument();
  });

  it('shows field errors before calling the server', async () => {
    const user = userEvent.setup();
    renderWithProviders(<Register />, { route: '/register', path: '/register' });
    await fill(user, 'mismatch');
    await user.click(screen.getByRole('button', { name: 'Create account' }));
    expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
  });

  it('shows the server error', async () => {
    const user = userEvent.setup();
    renderWithProviders(<Register />, {
      route: '/register', path: '/register',
      mocks: [{ request: { query: REGISTER, variables: { input } }, result: { errors: [new GraphQLError('already exists')] } }],
    });
    await fill(user);
    await user.click(screen.getByRole('button', { name: 'Create account' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('already exists');
  });

  it('redirects when already signed in', async () => {
    signIn();
    renderWithProviders(<Register />, { route: '/register', path: '/register', extraRoutes: dashboard });
    expect(await screen.findByText('Dashboard page')).toBeInTheDocument();
  });
});

describe('ProtectedRoute', () => {
  it('sends anonymous users to the login page', () => {
    renderWithProviders(<ProtectedRoute><p>secret</p></ProtectedRoute>, {
      route: '/dashboard', path: '/dashboard', extraRoutes: <Route path="/login" element={<p>Login page</p>} />,
    });
    expect(screen.getByText('Login page')).toBeInTheDocument();
  });
  it('renders children for signed-in users', () => {
    signIn();
    renderWithProviders(<ProtectedRoute><p>secret</p></ProtectedRoute>);
    expect(screen.getByText('secret')).toBeInTheDocument();
  });
});

describe('AuthContext', () => {
  const Probe = () => {
    const { user, logout, updateUser } = useAuthHook();
    return (
      <div>
        <span>{user ? user.name : 'anonymous'}</span>
        <button onClick={logout}>logout</button>
        <button onClick={() => updateUser({ ...alice, name: 'Renamed' })}>rename</button>
      </div>
    );
  };
  const mount = () =>
    render(<MockedProvider><AuthProvider><Probe /></AuthProvider></MockedProvider>);

  it('restores, updates and clears the session', async () => {
    signIn();
    const user = userEvent.setup();
    mount();
    expect(screen.getByText('Alice')).toBeInTheDocument();
    await user.click(screen.getByText('rename'));
    expect(screen.getByText('Renamed')).toBeInTheDocument();
    expect(JSON.parse(localStorage.getItem('user')!).name).toBe('Renamed');
    await user.click(screen.getByText('logout'));
    expect(screen.getByText('anonymous')).toBeInTheDocument();
    expect(localStorage.getItem('accessToken')).toBeNull();
  });

  it('logs out when the client reports an expired session', () => {
    signIn();
    mount();
    act(() => {
      window.dispatchEvent(new Event(LOGOUT_EVENT));
    });
    expect(screen.getByText('anonymous')).toBeInTheDocument();
  });

  it('refreshes the token periodically while signed in', () => {
    vi.useFakeTimers();
    const spy = vi.spyOn(apollo, 'refreshAccessToken').mockResolvedValue('x');
    signIn();
    const { unmount } = mount();
    act(() => {
      vi.advanceTimersByTime(REFRESH_INTERVAL_MS + 1);
    });
    expect(spy).toHaveBeenCalled();
    unmount();
    spy.mockRestore();
    vi.useRealTimers();
  });

  it('throws outside the provider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const Bad = () => {
      useAuth();
      return null;
    };
    expect(() => render(<Bad />)).toThrow('useAuth must be used within an AuthProvider');
    spy.mockRestore();
  });

  it('ignores a stored user without a token', async () => {
    localStorage.setItem('user', JSON.stringify(alice));
    mount();
    await waitFor(() => expect(screen.getByText('anonymous')).toBeInTheDocument());
  });
});
