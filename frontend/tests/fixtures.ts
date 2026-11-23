import type { Task, User } from '../src/types';

export const alice: User & { __typename: 'User' } = {
  __typename: 'User', id: 'u1', email: 'alice@example.com', name: 'Alice', avatar: null, createdAt: '2029-01-01T00:00:00Z',
};
export const bob: User & { __typename: 'User' } = {
  __typename: 'User', id: 'u2', email: 'bob@example.com', name: 'Bob', avatar: 'https://img.test/bob.png', createdAt: '2029-01-01T00:00:00Z',
};

export const makeTask = (over: Partial<Task> = {}): Task & { __typename: 'Task' } => ({
  __typename: 'Task',
  id: 't1',
  title: 'Write tests',
  description: 'Cover the board',
  status: 'TODO',
  priority: 'HIGH',
  dueDate: null,
  createdAt: '2030-01-01T00:00:00Z',
  updatedAt: '2030-01-01T00:00:00Z',
  createdBy: alice,
  assignedTo: null,
  ...over,
} as Task & { __typename: 'Task' });

export const signIn = (user: User = alice) => {
  localStorage.setItem('accessToken', 'access');
  localStorage.setItem('refreshToken', 'refresh');
  localStorage.setItem('user', JSON.stringify(user));
};
