import React from 'react';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../../tests/test-utils';
import TaskAssignment from './TaskAssignment';
import { User } from '../../types/user';

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

describe('TaskAssignment Component', () => {
  const mockOnClose = jest.fn();
  const mockOnAssign = jest.fn();
  const defaultProps = {
    open: true,
    onClose: mockOnClose,
    onAssign: mockOnAssign,
    taskId: '123',
    taskTitle: 'Test Task',
    users: mockUsers,
    loading: false,
  };

  beforeEach(() => {
    mockOnClose.mockClear();
    mockOnAssign.mockClear();
  });

  it('renders dialog with task title and user selection', () => {
    render(<TaskAssignment {...defaultProps} />);

    expect(screen.getByText('Assign Task')).toBeInTheDocument();
    expect(screen.getByText('Test Task')).toBeInTheDocument();
    expect(screen.getByLabelText('Select Assignee')).toBeInTheDocument();
  });

  it('displays current assignee when provided', () => {
    render(
      <TaskAssignment 
        {...defaultProps} 
        currentAssignee={mockUsers[0]} 
      />
    );

    expect(screen.getByText('Current Assignee:')).toBeInTheDocument();
    expect(screen.getByText('John Doe (john@example.com)')).toBeInTheDocument();
  });

  it('populates user selection with provided users', async () => {
    render(<TaskAssignment {...defaultProps} />);

    const select = screen.getByLabelText('Select Assignee');
    fireEvent.mouseDown(select);

    expect(await screen.findByText('John Doe')).toBeInTheDocument();
    expect(await screen.findByText('Jane Smith')).toBeInTheDocument();
  });

  it('calls onAssign with selected user ID when assign button is clicked', async () => {
    const user = userEvent.setup();
    render(<TaskAssignment {...defaultProps} />);

    const select = screen.getByLabelText('Select Assignee');
    await user.click(select);

    const johnOption = await screen.findByText('John Doe');
    await user.click(johnOption);

    const assignButton = screen.getByText('Assign Task');
    await user.click(assignButton);

    expect(mockOnAssign).toHaveBeenCalledWith('1');
  });

  it('disables assign button when no user is selected', () => {
    render(<TaskAssignment {...defaultProps} />);

    const assignButton = screen.getByText('Assign Task');
    expect(assignButton).toBeDisabled();
  });

  it('enables assign button when a user is selected', async () => {
    const user = userEvent.setup();
    render(<TaskAssignment {...defaultProps} />);

    const select = screen.getByLabelText('Select Assignee');
    await user.click(select);

    const johnOption = await screen.findByText('John Doe');
    await user.click(johnOption);

    const assignButton = screen.getByText('Assign Task');
    expect(assignButton).not.toBeDisabled();
  });

  it('shows loading state when loading prop is true', () => {
    render(<TaskAssignment {...defaultProps} loading={true} />);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('calls onClose when cancel button is clicked', async () => {
    const user = userEvent.setup();
    render(<TaskAssignment {...defaultProps} />);

    const cancelButton = screen.getByText('Cancel');
    await user.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('resets selection when dialog is closed and reopened', async () => {
    const user = userEvent.setup();
    const { rerender } = render(<TaskAssignment {...defaultProps} />);

    // Select a user
    const select = screen.getByLabelText('Select Assignee');
    await user.click(select);
    const johnOption = await screen.findByText('John Doe');
    await user.click(johnOption);

    // Close dialog
    await user.click(screen.getByText('Cancel'));

    // Reopen with same props (simulating parent state reset)
    rerender(<TaskAssignment {...defaultProps} open={false} />);
    rerender(<TaskAssignment {...defaultProps} open={true} />);

    // Selection should be reset
    expect(screen.getByLabelText('Select Assignee')).toHaveValue('');
  });

  it('displays warning when no users are available', () => {
    render(<TaskAssignment {...defaultProps} users={[]} />);

    expect(screen.getByText('No users available for assignment.')).toBeInTheDocument();
  });
});