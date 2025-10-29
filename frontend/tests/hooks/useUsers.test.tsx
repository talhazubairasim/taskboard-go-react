import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useUsers, useUser, useUpdateUser } from './useUsers';
import { User } from '../types/user';

// Mock API functions
jest.mock('../services/userService', () => ({
  fetchUsers: jest.fn(),
  fetchUser: jest.fn(),
  updateUser: jest.fn(),
}));

import * as userService from '../services/userService';

const mockUsers: User[] = [
  {
    id: 1,
    name: 'John Doe',
    email: 'john@example.com',
    role: 'user',
    isActive: true,
    avatar: '',
    createdAt: new Date(),
    updatedAt: new Date(),
  },
  {
    id: 2,
    name: 'Jane Smith',
    email: 'jane@example.com',
    role: 'admin',
    isActive: true,
    avatar: '',
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

describe('User Hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useUsers', () => {
    it('fetches users successfully', async () => {
      (userService.fetchUsers as jest.Mock).mockResolvedValue(mockUsers);

      const { result } = renderHook(() => useUsers(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(mockUsers);
      expect(userService.fetchUsers).toHaveBeenCalledTimes(1);
    });

    it('handles fetch error', async () => {
      const errorMessage = 'Failed to fetch users';
      (userService.fetchUsers as jest.Mock).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useUsers(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toBeInstanceOf(Error);
      expect(result.current.error?.message).toBe(errorMessage);
    });
  });

  describe('useUser', () => {
    it('fetches single user successfully', async () => {
      const userId = 1;
      (userService.fetchUser as jest.Mock).mockResolvedValue(mockUsers[0]);

      const { result } = renderHook(() => useUser(userId), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(result.current.data).toEqual(mockUsers[0]);
      expect(userService.fetchUser).toHaveBeenCalledWith(userId);
    });

    it('does not fetch when userId is undefined', async () => {
      const { result } = renderHook(() => useUser(undefined), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.isLoading).toBe(false));

      expect(userService.fetchUser).not.toHaveBeenCalled();
      expect(result.current.data).toBeUndefined();
    });
  });

  describe('useUpdateUser', () => {
    it('updates user successfully', async () => {
      const updatedUser = { ...mockUsers[0], name: 'Updated Name' };
      (userService.updateUser as jest.Mock).mockResolvedValue(updatedUser);

      const { result } = renderHook(() => useUpdateUser(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(updatedUser);
      });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(userService.updateUser).toHaveBeenCalledWith(updatedUser);
      expect(result.current.data).toEqual(updatedUser);
    });

    it('handles update error', async () => {
      const errorMessage = 'Failed to update user';
      (userService.updateUser as jest.Mock).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useUpdateUser(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockUsers[0]);
      });

      await waitFor(() => expect(result.current.isError).toBe(true));

      expect(result.current.error).toBeInstanceOf(Error);
      expect(result.current.error?.message).toBe(errorMessage);
    });
  });
});