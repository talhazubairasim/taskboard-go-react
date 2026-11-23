import { useQuery } from '@apollo/client';
import { GET_USERS } from '../graphql/queries';
import type { User } from '../types';

export function useUsers() {
  const { data, loading, error } = useQuery<{ users: User[] }>(GET_USERS);
  return { users: data?.users ?? [], loading, error };
}
