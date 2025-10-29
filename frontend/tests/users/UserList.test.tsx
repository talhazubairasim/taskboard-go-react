import React from 'react';
import { screen, fireEvent, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../../../tests/test-utils';
import UserList from './UserList';
import { User } from '../../../types/user';

const mockUsers: User[] = [
  {
    id: 1,
    name: 'John Doe',
    email: 'john@example.com',
    role: 'user',
    isActive: true,
    avatar: '',
    createdAt: new Date('2023-01-01'),
    updatedAt: new Date('2023-01-02'),
  },
  {
    id: 2,
    name: 'Jane Smith',
    email: 'jane@example.com',
    role: 'admin',
    isActive: false,
    avatar: '',
    createdAt: new Date('2023-01-03'),
    updatedAt: new Date('2023-01-04'),
  },
  {
    id: 3,
    name: 'Bob Johnson',
    email: 'bob@example.com',
    role: 'manager',
    isActive: true,
    avatar: '',
    createdAt: new Date('2023-01-05'),
    updatedAt: new Date('2023-01-06'),
  },
];

describe('UserList Component', () => {
  const mockOnEditUser = jest.fn();
  const mockOnDeleteUser = jest.fn();
  const mockOnViewUser = jest.fn();

  const defaultProps = {
    users: mockUsers,
    loading: false,
    onEditUser: mockOnEditUser,
    onDeleteUser: mockOnDeleteUser,
    onViewUser: mockOnViewUser,
  };

  beforeEach(() => {
    mockOnEditUser.mockClear();
    mockOnDeleteUser.mockClear();
    mockOnViewUser.mockClear();
  });

  it('renders user table with all users', () => {
    render(<UserList {...defaultProps} />);

    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    expect(screen.getByText('Bob Johnson')).toBeInTheDocument();
  });

  it('displays user information correctly', () => {
    render(<UserList {...defaultProps} />);

    const johnRow = screen.getByText('John Doe').closest('tr');
    expect(within(johnRow!).getByText('john@example.com')).toBeInTheDocument();
    expect(within(johnRow!).getByText('user')).toBeInTheDocument();
    expect(within(johnRow!).getByText('Active')).toBeInTheDocument();
  });

  it('filters users based on search input', async () => {
    const user = userEvent.setup();
    render(<UserList {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search users by name, email, or role...');
    await user.type(searchInput, 'john');

    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.queryByText('Jane Smith')).not.toBeInTheDocument();
    expect(screen.queryByText('Bob Johnson')).not.toBeInTheDocument();
  });

  it('calls onViewUser when view button is clicked', async () => {
    const user = userEvent.setup();
    render(<UserList {...defaultProps} />);

    const viewButtons = screen.getAllByLabelText('View Profile');
    await user.click(viewButtons[0]);

    expect(mockOnViewUser).toHaveBeenCalledWith(mockUsers[0]);
  });

  it('calls onEditUser when edit button is clicked', async () => {
    const user = userEvent.setup();
    render(<UserList {...defaultProps} />);

    const editButtons = screen.getAllByLabelText('Edit User');
    await user.click(editButtons[0]);

    expect(mockOnEditUser).toHaveBeenCalledWith(mockUsers[0]);
  });

  it('calls onDeleteUser when delete button is clicked', async () => {
    const user = userEvent.setup();
    render(<UserList {...defaultProps} />);

    const deleteButtons = screen.getAllByLabelText('Delete User');
    await user.click(deleteButtons[0]);

    expect(mockOnDeleteUser).toHaveBeenCalledWith(1);
  });

  it('disables delete button for admin users', () => {
    render(<UserList {...defaultProps} />);

    const adminRow = screen.getByText('Jane Smith').closest('tr');
    const deleteButton = within(adminRow!).getByLabelText('Delete User');
    
    expect(deleteButton).toBeDisabled();
  });

  it('shows loading state when loading is true', () => {
    render(<UserList {...defaultProps} loading={true} />);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('displays error message when error is provided', () => {
    const errorMessage = 'Failed to load users';
    render(<UserList {...defaultProps} error={errorMessage} />);

    expect(screen.getByText(errorMessage)).toBeInTheDocument();
  });

  it('handles pagination correctly', () => {
    const manyUsers = Array.from({ length: 15 }, (_, i) => ({
      ...mockUsers[0],
      id: i + 1,
      name: `User ${i + 1}`,
      email: `user${i + 1}@example.com`,
    }));

    render(<UserList {...defaultProps} users={manyUsers} />);

    // Should show 10 users per page by default
    expect(screen.getByText('User 1')).toBeInTheDocument();
    expect(screen.getByText('User 10')).toBeInTheDocument();
    expect(screen.queryByText('User 11')).not.toBeInTheDocument();

    // Change rows per page
    const rowsPerPageSelect = screen.getByLabelText('Rows per page:');
    fireEvent.mouseDown(rowsPerPageSelect);
    const option25 = screen.getByText('25');
    fireEvent.click(option25);

    // Should now show more users
    expect(screen.getByText('User 15')).toBeInTheDocument();
  });

  it('displays empty state when no users match search', async () => {
    const user = userEvent.setup();
    render(<UserList {...defaultProps} />);

    const searchInput = screen.getByPlaceholderText('Search users by name, email, or role...');
    await user.type(searchInput, 'nonexistent');

    expect(screen.getByText('No users found matching your search.')).toBeInTheDocument();
  });

  it('shows user count information', () => {
    render(<UserList {...defaultProps} />);

    expect(screen.getByText(`Showing ${mockUsers.length} of ${mockUsers.length} users`)).toBeInTheDocument();
  });
});