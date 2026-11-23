import type { ApolloCache, Reference } from '@apollo/client';
import { TASK_FIELDS } from '../graphql/fragments';
import type { Task } from '../types';

/** Adds a task to every cached `tasks` list (no-op if it is already there). */
export function addTaskToCache(cache: ApolloCache<unknown>, task: Task) {
  const ref = cache.writeFragment({ data: task, fragment: TASK_FIELDS, fragmentName: 'TaskFields' });
  if (!ref) return;
  cache.modify<{ tasks: Reference[] }>({
    fields: {
      tasks(existing = [], { readField }) {
        const list = existing as Reference[];
        if (list.some((r) => readField('id', r) === task.id)) return list;
        return [ref, ...list];
      },
    },
  });
}

/** Removes a task from the cache entirely. */
export function removeTaskFromCache(cache: ApolloCache<unknown>, id: string) {
  cache.evict({ id: cache.identify({ __typename: 'Task', id }) });
  cache.gc();
}
