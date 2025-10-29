import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useTasks, useTask, useCreateTask, useUpdateTask, useDeleteTask } from './useTasks';
import { Task } from '../types/task';

// Mock API functions
jest.mock('../services/taskService', () => ({
  fetchTasks: jest.fn(),
  fetchTask: jest.fn(),
  createTask: jest.fn(),
  updateTask: jest.fn(),
  deleteTask: jest.fn(),
}));

import * as taskService from '../services/taskService';

const mockTasks: Task[] = [
  {
    id: 1,
    title: 'Test Task 1',
    description: 'Description 1',
    status: 'todo',
    priority: 'medium',
    dueDate: new Date(),
    assigneeId: 1,
    createdById: 1,
    createdAt: new Date(),
    updatedAt: new Date(),
  },
  {
    id: 2,
    title: 'Test Task 2',
    description: 'Description 2',
    status: 'in-progress',
    priority: 'high',
    dueDate: new Date(),
    assigneeId: 2,
    createdById: 1,
    createdAt: new Date(),
    updatedAt: new Date(),
  },
];

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('Task Hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useTasks', () => {
    it('fetches tasks successfully', async () => {
      (taskService.fetchTasks as jest.Mock).mockResolvedValue(mockTasks);

      const { result } = renderHook(() => useTasks(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(mockTasks);
      expect(taskService.fetchTasks).toHaveBeenCalledTimes(1);
    });

    it('handles fetch error', async () => {
      const errorMessage = 'Failed to fetch tasks';
      (taskService.fetchTasks as jest.Mock).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useTasks(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toBeInstanceOf(Error);
      expect(result.current.error?.message).toBe(errorMessage);
    });
  });

  describe('useTask', () => {
    it('fetches single task successfully', async () => {
      const taskId = 1;
      (taskService.fetchTask as jest.Mock).mockResolvedValue(mockTasks[0]);

      const { result } = renderHook(() => useTask(taskId), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(mockTasks[0]);
      expect(taskService.fetchTask).toHaveBeenCalledWith(taskId);
    });

    it('does not fetch when taskId is undefined', async () => {
      const { result } = renderHook(() => useTask(undefined), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(taskService.fetchTask).not.toHaveBeenCalled();
      expect(result.current.data).toBeUndefined();
    });
  });

  describe('useCreateTask', () => {
    it('creates task successfully', async () => {
      const newTask = {
        title: 'New Task',
        description: 'New Description',
        status: 'todo' as const,
        priority: 'medium' as const,
        assigneeId: 1,
      };
      
      (taskService.createTask as jest.Mock).mockResolvedValue({ ...newTask, id: 3 });

      const { result } = renderHook(() => useCreateTask(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(newTask);
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(taskService.createTask).toHaveBeenCalledWith(newTask);
      expect(result.current.data).toEqual({ ...newTask, id: 3 });
    });

    it('handles create error', async () => {
      const errorMessage = 'Failed to create task';
      (taskService.createTask as jest.Mock).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useCreateTask(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          title: 'New Task',
          description: 'New Description',
          status: 'todo',
          priority: 'medium',
          assigneeId: 1,
        });
      });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toBeInstanceOf(Error);
      expect(result.current.error?.message).toBe(errorMessage);
    });
  });

  describe('useUpdateTask', () => {
    it('updates task successfully', async () => {
      const updatedTask = { ...mockTasks[0], title: 'Updated Title' };
      (taskService.updateTask as jest.Mock).mockResolvedValue(updatedTask);

      const { result } = renderHook(() => useUpdateTask(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(updatedTask);
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(taskService.updateTask).toHaveBeenCalledWith(updatedTask);
      expect(result.current.data).toEqual(updatedTask);
    });
  });

  describe('useDeleteTask', () => {
    it('deletes task successfully', async () => {
      const taskId = 1;
      (taskService.deleteTask as jest.Mock).mockResolvedValue(undefined);

      const { result } = renderHook(() => useDeleteTask(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(taskId);
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(taskService.deleteTask).toHaveBeenCalledWith(taskId);
    });
  });
});