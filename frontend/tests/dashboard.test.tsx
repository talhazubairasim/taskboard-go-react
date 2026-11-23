import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { GraphQLError } from 'graphql';
import type { MockedResponse } from '@apollo/client/testing';
import { Dashboard } from '../src/pages/Dashboard';
import { GET_TASKS, GET_USERS } from '../src/graphql/queries';
import { ASSIGN_TASK, CREATE_TASK, DELETE_TASK, UNASSIGN_TASK, UPDATE_TASK } from '../src/graphql/mutations';
import { TASK_CREATED, TASK_DELETED, TASK_UPDATED } from '../src/graphql/subscriptions';
import { alice, bob, makeTask, signIn } from './fixtures';
import { renderWithProviders } from './render';

const t1 = makeTask({ id: 't1', title: 'Mine', priority: 'HIGH' });
const t2 = makeTask({ id: 't2', title: 'Theirs', priority: 'LOW', createdBy: bob, description: 'needle in here' });

const tasksMock = (tasks = [t1, t2]): MockedResponse => ({ request: { query: GET_TASKS }, result: { data: { tasks } } });
const usersMock: MockedResponse = { request: { query: GET_USERS }, result: { data: { users: [alice, bob] } } };
// Subscriptions that never fire keep the "live" indicator on.
const quietSubs: MockedResponse[] = [
  { request: { query: TASK_CREATED }, result: { data: undefined }, delay: 1e9 },
  { request: { query: TASK_UPDATED }, result: { data: undefined }, delay: 1e9 },
  { request: { query: TASK_DELETED }, result: { data: undefined }, delay: 1e9 },
];

function setup(extra: MockedResponse[] = [], base: MockedResponse[] = [tasksMock(), usersMock, ...quietSubs]) {
  signIn();
  return renderWithProviders(<Dashboard />, { mocks: [...base, ...extra] });
}

const column = (name: string) => screen.getByRole('region', { name });

describe('Dashboard', () => {
  it('shows the board, greeting and live status', async () => {
    setup();
    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument();
    expect(await screen.findByText('Mine')).toBeInTheDocument();
    expect(screen.getByText('Welcome back, Alice!')).toBeInTheDocument();
    expect(screen.getByText('Live updates on')).toBeInTheDocument();
    expect(within(column('To Do')).getAllByRole('article')).toHaveLength(2);
  });

  it('filters by text, priority and ownership', async () => {
    const user = userEvent.setup();
    setup();
    await screen.findByText('Mine');
    await user.type(screen.getByLabelText('Search tasks'), 'needle');
    expect(screen.queryByText('Mine')).not.toBeInTheDocument();
    expect(screen.getByText('Theirs')).toBeInTheDocument();
    await user.clear(screen.getByLabelText('Search tasks'));
    await user.selectOptions(screen.getByLabelText('Filter by priority'), 'HIGH');
    expect(screen.queryByText('Theirs')).not.toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText('Filter by priority'), '');
    await user.click(screen.getByLabelText('Only my tasks'));
    expect(screen.getByText('Mine')).toBeInTheDocument();
    expect(screen.queryByText('Theirs')).not.toBeInTheDocument();
  });

  it('creates a task', async () => {
    const user = userEvent.setup();
    const created = makeTask({ id: 't3', title: 'Brand new', priority: 'MEDIUM', description: null });
    setup([{
      request: { query: CREATE_TASK, variables: { input: { title: 'Brand new', description: null, status: 'TODO', priority: 'MEDIUM', assignedToId: null, dueDate: null } } },
      result: { data: { createTask: created } },
    }]);
    await screen.findByText('Mine');
    await user.click(screen.getByText('+ Create Task'));
    const dialog = screen.getByRole('dialog', { name: 'Create New Task' });
    await user.type(within(dialog).getByLabelText('Title'), 'Brand new');
    await user.click(within(dialog).getByRole('button', { name: 'Create task' }));
    expect(await screen.findByText('Brand new')).toBeInTheDocument();
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('moves a card between columns', async () => {
    const user = userEvent.setup();
    setup([{
      request: { query: UPDATE_TASK, variables: { id: 't1', input: { status: 'DONE' } } },
      result: { data: { updateTask: { ...t1, status: 'DONE' } } },
    }]);
    await screen.findByText('Mine');
    await user.selectOptions(screen.getByLabelText('Move Mine'), 'DONE');
    await waitFor(() => expect(within(column('Done')).getByText('Mine')).toBeInTheDocument());
  });

  it('edits a task', async () => {
    const user = userEvent.setup();
    setup([{
      request: { query: UPDATE_TASK, variables: { id: 't1', input: { title: 'Mine (edited)', description: 'Cover the board', status: 'TODO', priority: 'HIGH', assignedToId: '', dueDate: null } } },
      result: { data: { updateTask: { ...t1, title: 'Mine (edited)' } } },
    }]);
    await screen.findByText('Mine');
    const card = screen.getByRole('article', { name: 'Mine' });
    await user.click(within(card).getByText('Edit'));
    const dialog = screen.getByRole('dialog', { name: 'Edit Task' });
    await user.type(within(dialog).getByLabelText('Title'), ' (edited)');
    await user.click(within(dialog).getByRole('button', { name: 'Save changes' }));
    expect(await screen.findByText('Mine (edited)')).toBeInTheDocument();
  });

  it('assigns and unassigns', async () => {
    const user = userEvent.setup();
    setup([
      { request: { query: ASSIGN_TASK, variables: { taskId: 't1', userId: bob.id } }, result: { data: { assignTask: { ...t1, assignedTo: bob } } } },
      { request: { query: UNASSIGN_TASK, variables: { taskId: 't1' } }, result: { data: { unassignTask: { ...t1, assignedTo: null } } } },
    ]);
    await screen.findByText('Mine');
    await user.click(within(screen.getByRole('article', { name: 'Mine' })).getByText('Assign'));
    await user.click(within(screen.getByRole('dialog')).getByText('Bob'));
    expect(await screen.findByTitle('Assigned to Bob')).toBeInTheDocument();
    await user.click(within(screen.getByRole('article', { name: 'Mine' })).getByText('Assign'));
    await user.click(screen.getByText('Remove assignee'));
    await waitFor(() => expect(screen.queryByTitle('Assigned to Bob')).not.toBeInTheDocument());
  });

  it('deletes a task after confirmation', async () => {
    const user = userEvent.setup();
    const confirm = vi.spyOn(window, 'confirm');
    setup([{ request: { query: DELETE_TASK, variables: { id: 't1' } }, result: { data: { deleteTask: true } } }]);
    await screen.findByText('Mine');
    confirm.mockReturnValueOnce(false);
    await user.click(within(screen.getByRole('article', { name: 'Mine' })).getByText('Delete'));
    expect(screen.getByText('Mine')).toBeInTheDocument();
    confirm.mockReturnValueOnce(true);
    await user.click(within(screen.getByRole('article', { name: 'Mine' })).getByText('Delete'));
    await waitFor(() => expect(screen.queryByText('Mine')).not.toBeInTheDocument());
    confirm.mockRestore();
  });

  it('reports failed actions and lets the user dismiss them', async () => {
    const user = userEvent.setup();
    setup([{
      request: { query: UPDATE_TASK, variables: { id: 't1', input: { status: 'REVIEW' } } },
      result: { errors: [new GraphQLError('you do not have permission to do that')] },
    }]);
    await screen.findByText('Mine');
    await user.selectOptions(screen.getByLabelText('Move Mine'), 'REVIEW');
    expect(await screen.findByRole('alert')).toHaveTextContent('you do not have permission');
    await user.click(screen.getByLabelText('Dismiss'));
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('shows load errors and paused live updates', async () => {
    setup([], [
      { request: { query: GET_TASKS }, result: { errors: [new GraphQLError('internal server error')] } },
      usersMock,
    ]);
    expect(await screen.findByRole('alert')).toHaveTextContent('internal server error');
    expect(await screen.findByText('Live updates paused')).toBeInTheDocument();
  });

  it('applies real-time creates and deletes from other users', async () => {
    const live = makeTask({ id: 't9', title: 'From a teammate', createdBy: bob });
    setup([], [
      tasksMock(),
      usersMock,
      { request: { query: TASK_CREATED }, result: { data: { taskCreated: live } }, delay: 50 },
      { request: { query: TASK_UPDATED }, result: { data: { taskUpdated: { ...t2, status: 'REVIEW' } } }, delay: 50 },
      { request: { query: TASK_DELETED }, result: { data: { taskDeleted: 't1' } }, delay: 50 },
    ]);
    expect(await screen.findByText('From a teammate')).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText('Mine')).not.toBeInTheDocument());
    await waitFor(() => expect(within(column('Review')).getByText('Theirs')).toBeInTheDocument());
  });

  it('renders nothing without a user', () => {
    const { container } = renderWithProviders(<Dashboard />, { mocks: [tasksMock(), usersMock, ...quietSubs] });
    expect(container).toBeEmptyDOMElement();
  });
});
