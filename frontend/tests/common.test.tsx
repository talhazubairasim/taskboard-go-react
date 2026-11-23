import { render, screen, fireEvent } from '@testing-library/react';
import { Avatar } from '../src/components/common/Avatar';
import { PriorityBadge } from '../src/components/common/PriorityBadge';
import { Modal } from '../src/components/common/Modal';
import { Spinner } from '../src/components/common/Spinner';
import { ErrorBoundary } from '../src/components/common/ErrorBoundary';

describe('common components', () => {
  it('Avatar shows an image or the initial', () => {
    const { rerender } = render(<Avatar name="bob" src="https://x/y.png" />);
    expect(screen.getByRole('img', { name: 'bob' })).toHaveAttribute('src', 'https://x/y.png');
    rerender(<Avatar name="bob" size="lg" />);
    expect(screen.getByLabelText('bob')).toHaveTextContent('B');
    rerender(<Avatar name=" " size="sm" />);
    expect(screen.getByText('?')).toBeInTheDocument();
  });

  it('PriorityBadge prints a readable label', () => {
    render(<PriorityBadge priority="URGENT" />);
    expect(screen.getByText('Urgent')).toBeInTheDocument();
  });

  it('Modal closes on Escape, backdrop and the close button, not on inner clicks', () => {
    const onClose = vi.fn();
    render(<Modal title="Hello" onClose={onClose}><p>Body</p></Modal>);
    fireEvent.click(screen.getByText('Body'));
    expect(onClose).not.toHaveBeenCalled();
    fireEvent.keyDown(window, { key: 'Enter' });
    expect(onClose).not.toHaveBeenCalled();
    fireEvent.keyDown(window, { key: 'Escape' });
    fireEvent.click(screen.getByLabelText('Close'));
    fireEvent.click(screen.getByRole('dialog').parentElement!);
    expect(onClose).toHaveBeenCalledTimes(3);
  });

  it('Spinner is announced as loading', () => {
    render(<Spinner />);
    expect(screen.getByRole('status', { name: 'Loading' })).toBeInTheDocument();
  });

  it('ErrorBoundary shows a fallback when a child throws', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const assign = vi.fn();
    Object.defineProperty(window, 'location', { value: { ...window.location, assign }, writable: true });
    const Boom = () => {
      throw new Error('kaboom');
    };
    render(<ErrorBoundary><Boom /></ErrorBoundary>);
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('kaboom')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Reload TaskBoard'));
    expect(assign).toHaveBeenCalledWith('/');
    spy.mockRestore();
  });

  it('ErrorBoundary renders children when nothing fails', () => {
    render(<ErrorBoundary><p>fine</p></ErrorBoundary>);
    expect(screen.getByText('fine')).toBeInTheDocument();
  });
});
