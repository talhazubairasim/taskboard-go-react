import type { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { MockedProvider, type MockedResponse } from '@apollo/client/testing';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { InMemoryCache } from '@apollo/client';
import { AuthProvider } from '../src/contexts/AuthContext';

export function createCache() {
  return new InMemoryCache({
    typePolicies: {
      Query: { fields: { tasks: { keyArgs: ['filter'], merge: (_e, incoming) => incoming } } },
    },
  });
}

interface Options {
  mocks?: MockedResponse[];
  route?: string;
  path?: string;
  cache?: InMemoryCache;
  extraRoutes?: ReactElement;
}

/** Renders ui inside Apollo (mocked), auth and router providers. */
export function renderWithProviders(ui: ReactElement, { mocks = [], route = '/', path = '*', cache, extraRoutes }: Options = {}) {
  const c = cache ?? createCache();
  const result = render(
    <MockedProvider mocks={mocks} cache={c}>
      <AuthProvider>
        <MemoryRouter initialEntries={[route]} future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
          <Routes>
            <Route path={path} element={ui} />
            {extraRoutes}
          </Routes>
        </MemoryRouter>
      </AuthProvider>
    </MockedProvider>,
  );
  return { ...result, cache: c };
}
