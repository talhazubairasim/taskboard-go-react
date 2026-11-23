import { Component, type ErrorInfo, type ReactNode } from 'react';

interface State {
  error: Error | null;
}

/** Shows a friendly message instead of a blank page if a component crashes. */
export class ErrorBoundary extends Component<{ children: ReactNode }, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Unhandled UI error', error, info.componentStack);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex min-h-screen flex-col items-center justify-center gap-4 p-8 text-center">
          <h1 className="text-2xl font-bold text-gray-900">Something went wrong</h1>
          <p className="text-gray-600">{this.state.error.message}</p>
          <button
            className="rounded-md bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-700"
            onClick={() => window.location.assign('/')}
          >
            Reload TaskBoard
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
