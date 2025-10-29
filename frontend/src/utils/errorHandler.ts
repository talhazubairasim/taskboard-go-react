// Error types
export interface AppError extends Error {
  code?: string;
  statusCode?: number;
  details?: any;
}

export interface ErrorResponse {
  success: false;
  error: {
    message: string;
    code?: string;
    statusCode: number;
    details?: any;
    timestamp: string;
  };
}

// Custom error classes
export class ValidationError extends Error implements AppError {
  statusCode = 400;
  code = 'VALIDATION_ERROR';
  details: any;

  constructor(message: string, details?: any) {
    super(message);
    this.name = 'ValidationError';
    this.details = details;
  }
}

export class AuthenticationError extends Error implements AppError {
  statusCode = 401;
  code = 'AUTHENTICATION_ERROR';

  constructor(message: string = 'Authentication required') {
    super(message);
    this.name = 'AuthenticationError';
  }
}

export class AuthorizationError extends Error implements AppError {
  statusCode = 403;
  code = 'AUTHORIZATION_ERROR';

  constructor(message: string = 'Insufficient permissions') {
    super(message);
    this.name = 'AuthorizationError';
  }
}

export class NotFoundError extends Error implements AppError {
  statusCode = 404;
  code = 'NOT_FOUND_ERROR';

  constructor(resource: string = 'Resource') {
    super(`${resource} not found`);
    this.name = 'NotFoundError';
  }
}

export class ConflictError extends Error implements AppError {
  statusCode = 409;
  code = 'CONFLICT_ERROR';

  constructor(message: string = 'Resource already exists') {
    super(message);
    this.name = 'ConflictError';
  }
}

export class InternalServerError extends Error implements AppError {
  statusCode = 500;
  code = 'INTERNAL_SERVER_ERROR';

  constructor(message: string = 'Internal server error') {
    super(message);
    this.name = 'InternalServerError';
  }
}

// Error handler utility
export class ErrorHandler {
  static handle(error: unknown): ErrorResponse {
    console.error('Error caught by handler:', error);

    // Handle known error types
    if (error instanceof ValidationError) {
      return this.createErrorResponse(
        error.message,
        error.statusCode,
        error.code,
        error.details
      );
    }

    if (error instanceof AuthenticationError) {
      return this.createErrorResponse(
        error.message,
        error.statusCode,
        error.code
      );
    }

    if (error instanceof AuthorizationError) {
      return this.createErrorResponse(
        error.message,
        error.statusCode,
        error.code
      );
    }

    if (error instanceof NotFoundError) {
      return this.createErrorResponse(
        error.message,
        error.statusCode,
        error.code
      );
    }

    if (error instanceof ConflictError) {
      return this.createErrorResponse(
        error.message,
        error.statusCode,
        error.code
      );
    }

    if (error instanceof InternalServerError) {
      return this.createErrorResponse(
        error.message,
        error.statusCode,
        error.code
      );
    }

    // Handle generic errors
    if (error instanceof Error) {
      return this.createErrorResponse(
        error.message,
        500,
        'UNKNOWN_ERROR'
      );
    }

    // Handle unknown error types
    return this.createErrorResponse(
      'An unexpected error occurred',
      500,
      'UNKNOWN_ERROR'
    );
  }

  private static createErrorResponse(
    message: string,
    statusCode: number,
    code?: string,
    details?: any
  ): ErrorResponse {
    return {
      success: false,
      error: {
        message,
        code,
        statusCode,
        details,
        timestamp: new Date().toISOString(),
      },
    };
  }

  // Validation helper
  static validateRequiredFields(fields: Record<string, any>, required: string[]): void {
    const missing: string[] = [];

    required.forEach(field => {
      if (fields[field] === undefined || fields[field] === null || fields[field] === '') {
        missing.push(field);
      }
    });

    if (missing.length > 0) {
      throw new ValidationError(
        `Missing required fields: ${missing.join(', ')}`,
        { missingFields: missing }
      );
    }
  }

  // Async error handler for Express middleware
  static asyncHandler(fn: Function) {
    return (req: any, res: any, next: any) => {
      Promise.resolve(fn(req, res, next)).catch(next);
    };
  }
}

// React hook for error handling
import { useState, useCallback } from 'react';

export const useErrorHandler = () => {
  const [error, setError] = useState<string | null>(null);

  const handleError = useCallback((error: unknown) => {
    const errorResponse = ErrorHandler.handle(error);
    setError(errorResponse.error.message);
    
    // You can also log to external service here
    console.error('Application error:', errorResponse);
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    error,
    handleError,
    clearError,
  };
};

// Higher-order component for error boundaries
import React from 'react';

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{ error: Error }>;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error boundary caught an error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      const FallbackComponent = this.props.fallback;
      
      if (FallbackComponent) {
        return <FallbackComponent error={this.state.error!} />;
      }

      return (
        <div style={{ padding: '20px', textAlign: 'center' }}>
          <h2>Something went wrong</h2>
          <p>{this.state.error?.message}</p>
          <button onClick={() => this.setState({ hasError: false })}>
            Try again
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorHandler;