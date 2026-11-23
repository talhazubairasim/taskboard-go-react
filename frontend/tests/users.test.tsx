import { screen } from '@testing-library/react';
import { render } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { GraphQLError } from 'graphql';
import { UserList } from '../src/components/users/UserList';
import { UserProfile } from '../src/components/users/UserProfile';
import { Team } from '../src/pages/Team';
import { Profile } from '../src/pages/Profile';
import { GET_TASKS, GET_USERS } from '../src/graphql/queries';
import { UPDATE_PROFILE } from '../src/graphql/mutations';
import { alice, bob, makeTask, signIn } from './fixtures';
import { renderWithProviders } from './render';

describe('UserList', () => {
  const tasks = [
    makeTask({ id: '1', assignedTo: bob }),
    makeTask({ id: '2', assignedTo: bob, status: 'DONE' }),
    makeTask({ id: '3', assignedTo: null }),
  ];

  it('lists people with their open and done counts', () => {
    render(<UserList users={[alice, bob]} tasks={tasks} />);
    const bobRow = screen.getByText('bob@example.com').closest('tr')!;
    expect(bobRow).toHaveTextContent(/Bob.*bob@example\.com\s*1\s*1/);
    const aliceRow = screen.getByText('alice@example.com').closest('tr')!;
    expect(aliceRow).toHaveTextContent(/0\s*0$/);
  });

  it('searches and shows an empty state', async () => {
    const user = userEvent.setup();
    render(<UserList users={[alice, bob]} tasks={[]} />);
    await user.type(screen.getByLabelText('Search team'), 'ali');
    expect(screen.queryByText('Bob')).not.toBeInTheDocument();
    await user.type(screen.getByLabelText('Search team'), 'zzz');
    expect(screen.getByText('No team members found')).toBeInTheDocument();
  });
});

describe('Team page', () => {
  it('loads users and tasks', async () => {
    signIn();
    renderWithProviders(<Team />, {
      mocks: [
        { request: { query: GET_USERS }, result: { data: { users: [alice, bob] } } },
        { request: { query: GET_TASKS }, result: { data: { tasks: [makeTask({ assignedTo: bob })] } } },
      ],
    });
    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument();
    expect(await screen.findByText('bob@example.com')).toBeInTheDocument();
  });
});

describe('UserProfile', () => {
  it('saves name and avatar and updates the session', async () => {
    const user = userEvent.setup();
    signIn();
    renderWithProviders(<Profile />, {
      mocks: [{
        request: { query: UPDATE_PROFILE, variables: { name: 'Alicia', avatar: 'https://img.test/a.png' } },
        result: { data: { updateProfile: { ...alice, name: 'Alicia', avatar: 'https://img.test/a.png' } } },
      }],
    });
    await user.clear(screen.getByLabelText('Name'));
    await user.type(screen.getByLabelText('Name'), 'Alicia');
    await user.type(screen.getByLabelText('Avatar URL'), 'https://img.test/a.png');
    expect(screen.getByRole('img', { name: 'Alicia' })).toHaveAttribute('src', 'https://img.test/a.png');
    await user.click(screen.getByRole('button', { name: 'Save profile' }));
    expect(await screen.findByRole('status')).toHaveTextContent('Profile saved');
    expect(JSON.parse(localStorage.getItem('user')!).name).toBe('Alicia');
  });

  it('validates the name and shows server errors', async () => {
    const user = userEvent.setup();
    signIn();
    renderWithProviders(<UserProfile />, {
      mocks: [{
        request: { query: UPDATE_PROFILE, variables: { name: 'Alice', avatar: 'ftp://bad' } },
        result: { errors: [new GraphQLError('avatar must be an http(s) URL')] },
      }],
    });
    await user.clear(screen.getByLabelText('Name'));
    await user.click(screen.getByRole('button', { name: 'Save profile' }));
    expect(screen.getByRole('alert')).toHaveTextContent('Name is required');
    await user.type(screen.getByLabelText('Name'), 'Alice');
    await user.type(screen.getByLabelText('Avatar URL'), 'ftp://bad');
    await user.click(screen.getByRole('button', { name: 'Save profile' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('avatar must be an http(s) URL');
  });

  it('renders nothing when signed out', () => {
    const { container } = renderWithProviders(<UserProfile />);
    expect(container).toBeEmptyDOMElement();
  });
});
