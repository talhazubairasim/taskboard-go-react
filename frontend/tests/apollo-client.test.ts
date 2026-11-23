import { gql } from '@apollo/client';
import { createApolloClient, LOGOUT_EVENT, refreshAccessToken } from '../src/utils/apollo-client';
import { tokens } from '../src/utils/tokens';

const json = (body: unknown, status = 200) =>
  Promise.resolve(new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } }));

const ME = gql`query Me { me { id } }`;
const LOGIN = gql`mutation Login { login(input: { email: "a", password: "b" }) { token } }`;
const unauthenticated = { errors: [{ message: 'authentication required', extensions: { code: 'UNAUTHENTICATED' } }] };
const refreshed = { data: { refreshToken: { token: 'new-access', refreshToken: 'new-refresh' } } };

function opName(init?: RequestInit) {
  return JSON.parse(String(init?.body)).query as string;
}

describe('refreshAccessToken', () => {
  it('stores the new token pair', async () => {
    tokens.save('old', 'refresh');
    const fetchImpl = vi.fn(() => json(refreshed));
    await expect(refreshAccessToken('http://api/query', fetchImpl as unknown as typeof fetch)).resolves.toBe('new-access');
    expect(tokens.access()).toBe('new-access');
    expect(tokens.refresh()).toBe('new-refresh');
  });

  it('shares one request between concurrent callers', async () => {
    tokens.save('old', 'refresh');
    const fetchImpl = vi.fn(() => json(refreshed));
    const f = fetchImpl as unknown as typeof fetch;
    const [a, b] = await Promise.all([refreshAccessToken('u', f), refreshAccessToken('u', f)]);
    expect(a).toBe('new-access');
    expect(b).toBe('new-access');
    expect(fetchImpl).toHaveBeenCalledTimes(1);
  });

  it.each([
    ['no refresh token', () => tokens.clear(), () => json(refreshed)],
    ['rejected refresh', () => tokens.save('a', 'r'), () => json(unauthenticated)],
    ['network failure', () => tokens.save('a', 'r'), () => Promise.reject(new Error('offline'))],
  ])('logs out on %s', async (_name, prepare, impl) => {
    prepare();
    const onLogout = vi.fn();
    window.addEventListener(LOGOUT_EVENT, onLogout);
    await expect(refreshAccessToken('u', vi.fn(impl) as unknown as typeof fetch)).resolves.toBeNull();
    expect(onLogout).toHaveBeenCalled();
    expect(tokens.access()).toBeNull();
    window.removeEventListener(LOGOUT_EVENT, onLogout);
  });
});

describe('createApolloClient', () => {
  it('sends the bearer token', async () => {
    tokens.save('access', 'refresh');
    const fetchImpl = vi.fn((_url: string, init?: RequestInit) => {
      expect((init?.headers as Record<string, string>).authorization).toBe('Bearer access');
      return json({ data: { me: { id: '1', __typename: 'User' } } });
    });
    const client = createApolloClient({ fetchImpl: fetchImpl as unknown as typeof fetch });
    const res = await client.query({ query: ME });
    expect(res.data.me.id).toBe('1');
  });

  it('omits the header when signed out', async () => {
    const fetchImpl = vi.fn((_url: string, init?: RequestInit) => {
      expect((init?.headers as Record<string, string>).authorization).toBeUndefined();
      return json({ data: { me: { id: '1', __typename: 'User' } } });
    });
    const client = createApolloClient({ fetchImpl: fetchImpl as unknown as typeof fetch, subscriptions: false });
    await client.query({ query: ME });
  });

  it('refreshes and retries once on UNAUTHENTICATED', async () => {
    tokens.save('expired', 'refresh');
    const fetchImpl = vi.fn((_url: string, init?: RequestInit) => {
      if (opName(init).includes('refreshToken')) return json(refreshed);
      const auth = (init?.headers as Record<string, string>).authorization;
      return auth === 'Bearer new-access' ? json({ data: { me: { id: '9', __typename: 'User' } } }) : json(unauthenticated);
    });
    const client = createApolloClient({ fetchImpl: fetchImpl as unknown as typeof fetch, subscriptions: false });
    const res = await client.query({ query: ME });
    expect(res.data.me.id).toBe('9');
    expect(fetchImpl).toHaveBeenCalledTimes(3);
  });

  it('refreshes on an HTTP 401 for an expired token', async () => {
    tokens.save('expired', 'refresh');
    const fetchImpl = vi.fn((_url: string, init?: RequestInit) => {
      if (opName(init).includes('refreshToken')) return json(refreshed);
      const auth = (init?.headers as Record<string, string>).authorization;
      return auth === 'Bearer new-access'
        ? json({ data: { me: { id: '7', __typename: 'User' } } })
        : json({ error: 'invalid or expired token' }, 401);
    });
    const client = createApolloClient({ fetchImpl: fetchImpl as unknown as typeof fetch, subscriptions: false });
    const res = await client.query({ query: ME });
    expect(res.data.me.id).toBe('7');
  });

  it('gives up when the refresh fails', async () => {
    tokens.save('expired', 'refresh');
    const fetchImpl = vi.fn(() => json(unauthenticated));
    const client = createApolloClient({ fetchImpl: fetchImpl as unknown as typeof fetch, subscriptions: false });
    await expect(client.query({ query: ME })).rejects.toThrow(/session has expired/);
  });

  it('does not retry auth operations or other errors', async () => {
    const fetchImpl = vi.fn(() => json(unauthenticated));
    const client = createApolloClient({ fetchImpl: fetchImpl as unknown as typeof fetch, subscriptions: false });
    await expect(client.mutate({ mutation: LOGIN })).rejects.toThrow('authentication required');
    expect(fetchImpl).toHaveBeenCalledTimes(1);

    const other = vi.fn(() => json({ errors: [{ message: 'nope', extensions: { code: 'FORBIDDEN' } }] }));
    const c2 = createApolloClient({ fetchImpl: other as unknown as typeof fetch, subscriptions: false });
    await expect(c2.query({ query: ME })).rejects.toThrow('nope');
    expect(other).toHaveBeenCalledTimes(1);
  });
});
