import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MockedProvider } from '@apollo/client/testing';
import { MemoryRouter } from 'react-router-dom';
import App, { AppRoutes } from '../src/App';
import { AuthProvider } from '../src/contexts/AuthContext';
import { createApolloClient } from '../src/utils/apollo-client';
import { GET_TASKS, GET_USERS } from '../src/graphql/queries';
import { alice, bob, signIn } from './fixtures';

const renderAt = (path: string) =>
  render(
    <MockedProvider mocks={[
      { request: { query: GET_TASKS }, result: { data: { tasks: [] } }, maxUsageCount: 5 },
      { request: { query: GET_USERS }, result: { data: { users: [alice, bob] } }, maxUsageCount: 5 },
    ]}>
      <AuthProvider>
        <MemoryRouter initialEntries={[path]} future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
          <AppRoutes />
        </MemoryRouter>
      </AuthProvider>
    </MockedProvider>,
  );

describe('App routing', () => {
  it('sends anonymous visitors to sign in', () => {
    renderAt('/dashboard');
    expect(screen.getByText('Sign in to TaskBoard')).toBeInTheDocument();
  });

  it('shows the register page', () => {
    renderAt('/register');
    expect(screen.getByText('Create your account')).toBeInTheDocument();
  });

  it('navigates between board, team and profile, and logs out', async () => {
    const user = userEvent.setup();
    signIn();
    renderAt('/nowhere');
    expect(await screen.findByText('Welcome back, Alice!')).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'Main' })).toBeInTheDocument();
    await user.click(screen.getByRole('link', { name: 'Team' }));
    expect(await screen.findByRole('heading', { name: 'Team' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Team' })).toHaveClass('bg-indigo-50');
    await user.click(screen.getByRole('link', { name: 'Profile' }));
    expect(screen.getByRole('heading', { name: 'Your profile' })).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Logout' }));
    expect(await screen.findByText('Sign in to TaskBoard')).toBeInTheDocument();
  });

  it('header shows the avatar image when the user has one', () => {
    signIn(bob);
    renderAt('/profile');
    expect(screen.getAllByRole('img', { name: 'Bob' }).length).toBeGreaterThan(0);
  });

  it('mounts the full app with a real client', () => {
    window.history.pushState({}, '', '/login');
    render(<App client={createApolloClient({ subscriptions: false })} />);
    expect(screen.getByText('Sign in to TaskBoard')).toBeInTheDocument();
  });

  it('mounts with the default client', () => {
    window.history.pushState({}, '', '/register');
    render(<App />);
    expect(screen.getByText('Create your account')).toBeInTheDocument();
  });
});
