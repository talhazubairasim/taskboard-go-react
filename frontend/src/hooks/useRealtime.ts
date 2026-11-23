import { useSubscription } from '@apollo/client';
import { TASK_CREATED, TASK_DELETED, TASK_UPDATED } from '../graphql/subscriptions';
import type { Task } from '../types';
import { addTaskToCache, removeTaskFromCache } from './taskCache';

/**
 * Keeps the board in sync with other users' changes over WebSocket.
 * Updated tasks are normalised into the cache automatically; created and
 * deleted tasks are added to / removed from the cached lists here.
 */
export function useRealtime() {
  const created = useSubscription<{ taskCreated: Task }>(TASK_CREATED, {
    onData: ({ client, data }) => {
      if (data.data) addTaskToCache(client.cache, data.data.taskCreated);
    },
  });
  const updated = useSubscription<{ taskUpdated: Task }>(TASK_UPDATED);
  const deleted = useSubscription<{ taskDeleted: string }>(TASK_DELETED, {
    onData: ({ client, data }) => {
      if (data.data) removeTaskFromCache(client.cache, data.data.taskDeleted);
    },
  });
  const error = created.error || updated.error || deleted.error;
  return { connected: !error, error };
}
