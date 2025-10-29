import React from 'react';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '../../../tests/test-utils';
import UserProfile from './UserProfile';
import { User } from '../../../types/user';

const mockUser: User = {
  id: 1,
  name: 'John Doe',
  email: 'john@example.com',
  role: 'user',
  isActive: true,
  avatar: '',
  department: 'Engineering',
  position: 'Software Engineer',
  createdAt: new Date('2023-01-01T00:00:00Z'),
  updatedAt: new Date('2023-01-02T00:00:00Z'),
};

describe('UserProfile Component', () => {
  const mockOnUpdateUser = jest.fn();
  const defaultProps = {
    user: mockUser,
    onUpdateUser: mockOnUpdateUser,
    loading: false,
    isEditable: true,
  };

  beforeEach(() => {
    mockOnUpdateUser.mockClear();
  });

  it('renders user profile information correctly', () => {
    render(<UserProfile {...defaultProps} />);

    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('john@example.com')).toBeInTheDocument();
    expect(screen.getByText('user')).toBeInTheDocument();
    expect(screen.getByText('Active')).toBeInTheDocument();
  });

  it('displays edit button when isEditable is true', () => {
    render(<UserProfile {...defaultProps} />);

    expect(screen.getByText('Edit Profile')).toBeInTheDocument();
  });

  it('hides edit button when isEditable is false', () => {
    render(<UserProfile {...defaultProps} isEditable={false} />);

    expect(screen.queryByText('Edit Profile')).not.toBeInTheDocument();
  });

  it('enters edit mode when edit button is clicked', async () => {
    const user = userEvent.setup();
    render(<UserProfile {...defaultProps} />);

    const editButton = screen.getByText('Edit Profile');
    await user.click(editButton);

    expect(screen.getByText('Cancel')).toBeInTheDocument();
    expect(screen.getByText('Save Changes')).toBeInTheDocument();
    expect(screen.getByDisplayValue('John Doe')).toBeInTheDocument();
  });

  it('updates user information and calls onUpdateUser', async () => {
    const user = userEvent.setup();
    mockOnUpdateUser.mockResolvedValueOnce(undefined);

    render(<UserProfile {...defaultProps} />);

    // Enter edit mode
    const editButton = screen.getByText('Edit Profile');
    await user.click(editButton);

    // Change name
    const nameInput = screen.getByDisplayValue('John Doe');
    await user.clear(nameInput);
    await user.type(nameInput, 'John Updated');

    // Change role
    const roleSelect = screen.getByLabelText('Role');
    await user.click(roleSelect);
    const managerOption = await screen.findByText('Manager');
    await user.click(managerOption);

    // Save changes
    const saveButton = screen.getByText('Save Changes');
    await user.click(saveButton);

    await waitFor(() => {
      expect(mockOnUpdateUser).toHaveBeenCalledWith({
        ...mockUser,
        name: 'John Updated',
        role: 'manager',
      });
    });
  });

  it('cancels edit mode without saving changes', async () => {
    const user = userEvent.setup();
    render(<UserProfile {...defaultProps} />);

    // Enter edit mode
    const editButton = screen.getByText('Edit Profile');
    await user.click(editButton);

    // Change name
    const nameInput = screen.getByDisplayValue('John Doe');
    await user.clear(nameInput);
    await user.type(nameInput, 'John Updated');

    // Cancel
    const cancelButton = screen.getByText('Cancel');
    await user.click(cancelButton);

    // Should revert to original name
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(mockOnUpdateUser).not.toHaveBeenCalled();
  });

  it('displays loading state during save', async () => {
    const user = userEvent.setup();
    mockOnUpdateUser.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

    render(<UserProfile {...defaultProps} loading={true} />);

    // Enter edit mode
    const editButton = screen.getByText('Edit Profile');
    await user.click(editButton);

    // Try to save
    const saveButton = screen.getByText('Save Changes');
    await user.click(saveButton);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('displays error message when update fails', async () => {
    const user = userEvent.setup();
    const errorMessage = 'Failed to update user';
    mockOnUpdateUser.mockRejectedValueOnce(new Error(errorMessage));

    render(<UserProfile {...defaultProps} />);

    // Enter edit mode and save
    const editButton = screen.getByText('Edit Profile');
    await user.click(editButton);
    const saveButton = screen.getByText('Save Changes');
    await user.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('displays success message when update succeeds', async () => {
    const user = userEvent.setup();
    mockOnUpdateUser.mockResolvedValueOnce(undefined);

    render(<UserProfile {...defaultProps} />);

    // Enter edit mode and save
    const editButton = screen.getByText('Edit Profile');
    await user.click(editButton);
    const saveButton = screen.getByText('Save Changes');
    await user.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText('Profile updated successfully!')).toBeInTheDocument();
    });
  });

  it('displays additional user information when available', () => {
    render(<UserProfile {...defaultProps} />);

    expect(screen.getByText('Engineering')).toBeInTheDocument();
    expect(screen.getByText('Software Engineer')).toBeInTheDocument();
  });

  it('formats dates correctly', () => {
    render(<UserProfile {...defaultProps} />);

    expect(screen.getByDisplayValue('1/1/2023, 12:00:00 AM')).toBeInTheDocument();
    expect(screen.getByDisplayValue('1/2/2023, 12:00:00 AM')).toBeInTheDocument();
  });

  it('applies correct chip colors for roles and status', () => {
    render(<UserProfile {...defaultProps} />);

    const userChip = screen.getByText('user');
    const activeChip = screen.getByText('Active');

    expect(userChip).toHaveClass('MuiChip-colorPrimary');
    expect(activeChip).toHaveClass('MuiChip-colorSuccess');
  });
});