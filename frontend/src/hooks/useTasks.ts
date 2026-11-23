import { useMutation, useQuery } from '@apollo/client';
import { GET_TASKS } from '../graphql/queries';
import {
  ASSIGN_TASK,
  CREATE_TASK,
  DELETE_TASK,
  UNASSIGN_TASK,
  UPDATE_TASK,
} from '../graphql/mutations';
import type { Priority, Task, TaskStatus } from '../types';
import { addTaskToCache, removeTaskFromCache } from './taskCache';

export interface TaskInput {
  title?: string;
  description?: string | null;
  status?: TaskStatus;
  priority?: Priority;
  assignedToId?: string | null;
  dueDate?: string | null;
}

/** Loads every task on the board. The board groups them by status itself. */
export function useTasks() {
  const { data, loading, error, refetch } = useQuery<{ tasks: Task[] }>(GET_TASKS);
  return { tasks: data?.tasks ?? [], loading, error, refetch };
}

/** All task mutations, with cache updates so the board reacts immediately. */
export function useTaskActions() {
  const [create, createState] = useMutation<{ createTask: Task }>(CREATE_TASK, {
    update: (cache, { data }) => data && addTaskToCache(cache, data.createTask),
  });
  const [update, updateState] = useMutation<{ updateTask: Task }>(UPDATE_TASK);
  const [remove] = useMutation(DELETE_TASK);
  const [assign] = useMutation<{ assignTask: Task }>(ASSIGN_TASK);
  const [unassign] = useMutation<{ unassignTask: Task }>(UNASSIGN_TASK);

  return {
    creating: createState.loading,
    updating: updateState.loading,
    createTask: (input: TaskInput) => create({ variables: { input } }),
    updateTask: (id: string, input: TaskInput) => update({ variables: { id, input } }),
    /** Moves a card to another column, updating the UI before the server answers. */
    moveTask: (task: Task, status: TaskStatus) =>
      update({
        variables: { id: task.id, input: { status } },
        optimisticResponse: { updateTask: { ...task, status, __typename: 'Task' } as Task },
      }),
    deleteTask: (id: string) =>
      remove({ variables: { id }, update: (cache) => removeTaskFromCache(cache, id) }),
    assignTask: (taskId: string, userId: string) => assign({ variables: { taskId, userId } }),
    unassignTask: (taskId: string) => unassign({ variables: { taskId } }),
  };
}
