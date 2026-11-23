import { ApolloError } from '@apollo/client';
import { GraphQLError } from 'graphql';
import { getErrorCode, getErrorMessage } from '../src/utils/errors';
import { formatDate, isOverdue, toDueDate } from '../src/utils/format';
import { tokens } from '../src/utils/tokens';
import { validateRegistration } from '../src/utils/validation';
import { makeTask } from './fixtures';

describe('errors', () => {
  it('prefers the first GraphQL error message and exposes its code', () => {
    const err = new ApolloError({ graphQLErrors: [new GraphQLError('nope', { extensions: { code: 'FORBIDDEN' } })] });
    expect(getErrorMessage(err)).toBe('nope');
    expect(getErrorCode(err)).toBe('FORBIDDEN');
  });
  it('describes network errors', () => {
    const err = new ApolloError({ networkError: new Error('down') });
    expect(getErrorMessage(err)).toMatch(/Cannot reach the server/);
    expect(getErrorCode(err)).toBeUndefined();
  });
  it('falls back for plain and unknown errors', () => {
    expect(getErrorMessage(new Error('plain'))).toBe('plain');
    expect(getErrorMessage('weird', 'fallback')).toBe('fallback');
    expect(getErrorMessage(new ApolloError({}), 'fb')).toBe('fb');
  });
});

describe('format', () => {
  it('formats dates in UTC', () => {
    expect(formatDate('2030-01-02T00:00:00Z')).toMatch(/2030/);
  });
  it('detects overdue tasks', () => {
    const now = new Date('2030-06-01T00:00:00Z');
    expect(isOverdue(makeTask({ dueDate: '2030-05-01T00:00:00Z' }), now)).toBe(true);
    expect(isOverdue(makeTask({ dueDate: '2030-05-01T00:00:00Z', status: 'DONE' }), now)).toBe(false);
    expect(isOverdue(makeTask({ dueDate: '2030-07-01T00:00:00Z' }), now)).toBe(false);
    expect(isOverdue(makeTask({ dueDate: null }), now)).toBe(false);
    expect(isOverdue(makeTask({ dueDate: '2000-01-01T00:00:00Z' }))).toBe(true);
  });
  it('converts date inputs to RFC 3339', () => {
    expect(toDueDate('2030-03-04')).toBe('2030-03-04T00:00:00Z');
    expect(toDueDate('')).toBeNull();
  });
});

describe('tokens', () => {
  it('saves, reads and clears the session', () => {
    tokens.save('a', 'r', { id: '1', email: 'e', name: 'n' });
    expect(tokens.access()).toBe('a');
    expect(tokens.refresh()).toBe('r');
    expect(tokens.user()?.name).toBe('n');
    tokens.save('a2', 'r2');
    expect(tokens.user()?.name).toBe('n');
    tokens.saveUser({ id: '1', email: 'e', name: 'm' });
    expect(tokens.user()?.name).toBe('m');
    tokens.clear();
    expect(tokens.access()).toBeNull();
    expect(tokens.user()).toBeNull();
  });
  it('ignores a corrupt stored user', () => {
    localStorage.setItem('user', '{bad');
    expect(tokens.user()).toBeNull();
  });
});

describe('validateRegistration', () => {
  const ok = { name: 'A', email: 'a@b.co', password: 'Str0ng!Pass', confirmPassword: 'Str0ng!Pass' };
  it('accepts a valid form', () => {
    expect(validateRegistration(ok)).toEqual({});
  });
  it.each([
    [{ name: ' ' }, 'name', 'Name is required'],
    [{ email: '' }, 'email', 'Email is required'],
    [{ email: 'nope' }, 'email', 'Invalid email format'],
    [{ password: '', confirmPassword: '' }, 'password', 'Password is required'],
    [{ password: 'Sh0rt!', confirmPassword: 'Sh0rt!' }, 'password', 'at least 8'],
    [{ password: 'UPPER123!', confirmPassword: 'UPPER123!' }, 'password', 'lowercase'],
    [{ password: 'lower123!', confirmPassword: 'lower123!' }, 'password', 'uppercase'],
    [{ password: 'NoNumber!', confirmPassword: 'NoNumber!' }, 'password', 'number'],
    [{ password: 'NoSpecial1', confirmPassword: 'NoSpecial1' }, 'password', 'special'],
    [{ confirmPassword: 'different' }, 'confirmPassword', 'do not match'],
  ])('rejects %o', (over, field, msg) => {
    expect(validateRegistration({ ...ok, ...over })[field]).toMatch(msg);
  });
});
