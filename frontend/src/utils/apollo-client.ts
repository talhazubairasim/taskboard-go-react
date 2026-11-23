import { ApolloClient, ApolloLink, HttpLink, InMemoryCache, Observable, split } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { getMainDefinition } from '@apollo/client/utilities';
import { createClient } from 'graphql-ws';
import { tokens } from './tokens';

export const HTTP_URL = import.meta.env.VITE_GRAPHQL_HTTP_URL || 'http://localhost:8080/query';
export const WS_URL = import.meta.env.VITE_GRAPHQL_WS_URL || 'ws://localhost:8080/query';

/** Fired when the session can no longer be refreshed; AuthContext logs the user out. */
export const LOGOUT_EVENT = 'taskboard:logout';

const REFRESH_MUTATION = `mutation RefreshToken($refreshToken: String!) {
  refreshToken(refreshToken: $refreshToken) { token refreshToken }
}`;

let inFlight: Promise<string | null> | null = null;

/**
 * Exchanges the stored refresh token for a new token pair. Concurrent callers
 * share one request. Resolves to the new access token, or null if the session
 * has expired (in which case tokens are cleared and LOGOUT_EVENT is fired).
 */
export function refreshAccessToken(url = HTTP_URL, fetchImpl: typeof fetch = fetch): Promise<string | null> {
  if (inFlight) return inFlight;
  const refreshToken = tokens.refresh();
  const run = async (): Promise<string | null> => {
    if (!refreshToken) return null;
    try {
      const res = await fetchImpl(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: REFRESH_MUTATION, variables: { refreshToken } }),
      });
      const body = await res.json();
      const pair = body?.data?.refreshToken;
      if (!pair?.token) return null;
      tokens.save(pair.token, pair.refreshToken);
      return pair.token as string;
    } catch {
      return null;
    }
  };
  inFlight = run().then((token) => {
    inFlight = null;
    if (!token) {
      tokens.clear();
      window.dispatchEvent(new Event(LOGOUT_EVENT));
    }
    return token;
  });
  return inFlight;
}

const AUTH_OPERATIONS = new Set(['Login', 'Register', 'RefreshToken']);

function isUnauthenticated(
  graphQLErrors: readonly { extensions?: Record<string, unknown> }[] | undefined,
  networkError: unknown,
): boolean {
  if (graphQLErrors?.some((e) => e.extensions?.code === 'UNAUTHENTICATED')) return true;
  const status = (networkError as { statusCode?: number } | null | undefined)?.statusCode;
  return status === 401;
}

interface ClientOptions {
  httpUrl?: string;
  wsUrl?: string;
  fetchImpl?: typeof fetch;
  /** Set to false in tests to skip the WebSocket link. */
  subscriptions?: boolean;
}

export function createApolloClient({
  httpUrl = HTTP_URL,
  wsUrl = WS_URL,
  fetchImpl,
  subscriptions = true,
}: ClientOptions = {}) {
  const httpLink = new HttpLink({ uri: httpUrl, fetch: fetchImpl });

  const authLink = setContext((_, { headers }) => {
    const token = tokens.access();
    return { headers: { ...headers, ...(token ? { authorization: `Bearer ${token}` } : {}) } };
  });

  // On UNAUTHENTICATED (or HTTP 401 for an expired token) refresh once and retry.
  const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
    if (AUTH_OPERATIONS.has(operation.operationName) || operation.getContext().retried) return;
    if (!isUnauthenticated(graphQLErrors, networkError)) return;
    return new Observable((observer) => {
      refreshAccessToken(httpUrl, fetchImpl).then((token) => {
        if (!token) {
          observer.error(new Error('Your session has expired. Please sign in again.'));
          return;
        }
        operation.setContext(({ headers = {} }) => ({
          retried: true,
          headers: { ...headers, authorization: `Bearer ${token}` },
        }));
        forward(operation).subscribe(observer);
      });
    });
  });

  let link: ApolloLink = ApolloLink.from([errorLink, authLink, httpLink]);

  if (subscriptions) {
    const wsLink = new GraphQLWsLink(
      createClient({
        url: wsUrl,
        lazy: true,
        retryAttempts: Infinity,
        connectionParams: () => {
          const token = tokens.access();
          return token ? { Authorization: `Bearer ${token}` } : {};
        },
      }),
    );
    link = split(
      ({ query }) => {
        const def = getMainDefinition(query);
        return def.kind === 'OperationDefinition' && def.operation === 'subscription';
      },
      wsLink,
      link,
    );
  }

  return new ApolloClient({
    link,
    cache: new InMemoryCache({
      typePolicies: {
        Query: {
          fields: {
            tasks: { keyArgs: ['filter'], merge: (_existing, incoming) => incoming },
          },
        },
      },
    }),
    defaultOptions: {
      watchQuery: { fetchPolicy: 'cache-and-network' },
    },
  });
}

export const client = createApolloClient();
