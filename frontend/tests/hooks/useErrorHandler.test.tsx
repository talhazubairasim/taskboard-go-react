import { renderHook, act } from '@testing-library/react';
import { useErrorHandler } from './useErrorHandler';

describe('useErrorHandler Hook', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('initializes with null error', () => {
    const { result } = renderHook(() => useErrorHandler());

    expect(result.current.error).toBeNull();
  });

  it('sets error when handleError is called', () => {
    const { result } = renderHook(() => useErrorHandler());

    act(() => {
      result.current.handleError(new Error('Test error'));
    });

    expect(result.current.error).toBe('Test error');
  });

  it('clears error when clearError is called', () => {
    const { result } = renderHook(() => useErrorHandler());

    act(() => {
      result.current.handleError(new Error('Test error'));
    });

    expect(result.current.error).toBe('Test error');

    act(() => {
      result.current.clearError();
    });

    expect(result.current.error).toBeNull();
  });

  it('handles string errors', () => {
    const { result } = renderHook(() => useErrorHandler());

    act(() => {
      result.current.handleError('String error');
    });

    expect(result.current.error).toBe('An unexpected error occurred');
  });

  it('handles custom error types', () => {
    const { result } = renderHook(() => useErrorHandler());

    class CustomError extends Error {
      constructor(message: string) {
        super(message);
        this.name = 'CustomError';
      }
    }

    act(() => {
      result.current.handleError(new CustomError('Custom error message'));
    });

    expect(result.current.error).toBe('Custom error message');
  });

  it('handles null/undefined errors', () => {
    const { result } = renderHook(() => useErrorHandler());

    act(() => {
      result.current.handleError(null);
    });

    expect(result.current.error).toBe('An unexpected error occurred');

    act(() => {
      result.current.clearError();
      result.current.handleError(undefined);
    });

    expect(result.current.error).toBe('An unexpected error occurred');
  });
});