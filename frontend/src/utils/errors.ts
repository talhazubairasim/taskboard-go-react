import { ApolloError } from '@apollo/client';

/** Turns an Apollo or network error into a short message for the UI. */
export function getErrorMessage(error: unknown, fallback = 'Something went wrong. Please try again.'): string {
  if (error instanceof ApolloError) {
    if (error.graphQLErrors.length > 0) return error.graphQLErrors[0].message;
    if (error.networkError) return 'Cannot reach the server. Check your connection.';
  }
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

/** Returns the GraphQL error code (extensions.code) of the first error, if any. */
export function getErrorCode(error: unknown): string | undefined {
  if (error instanceof ApolloError && error.graphQLErrors.length > 0) {
    return error.graphQLErrors[0].extensions?.code as string | undefined;
  }
  return undefined;
}
