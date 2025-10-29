import React from 'react';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { render } from '../../tests/test-utils';
import Sidebar from './Sidebar';

// Mock useLocation
const mockUseLocation = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useLocation: () => mockUseLocation(),
}));

describe('Sidebar Component', () => {
  const mockOnDrawerToggle = jest.fn();
  const defaultProps = {
    mobileOpen: false,
    onDrawerToggle: mockOnDrawerToggle,
  };

  beforeEach(() => {
    mockUseLocation.mockReturnValue({ pathname: '/' });
    mockOnDrawerToggle.mockClear();
  });

  it('renders all navigation items correctly', () => {
    render(<Sidebar {...defaultProps} />);

    expect(screen.getByText('TaskBoard')).toBeInTheDocument();
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Tasks')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('highlights the active menu item based on current route', () => {
    mockUseLocation.mockReturnValue({ pathname: '/tasks' });
    
    render(<Sidebar {...defaultProps} />);

    const tasksButton = screen.getByText('Tasks').closest('button');
    expect(tasksButton).toHaveClass('Mui-selected');

    const dashboardButton = screen.getByText('Dashboard').closest('button');
    expect(dashboardButton).not.toHaveClass('Mui-selected');
  });

  it('calls onDrawerToggle when menu item is clicked on mobile view', () => {
    // Mock mobile view by setting smaller window width
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 600,
    });

    render(<Sidebar {...defaultProps} mobileOpen={true} />);

    const dashboardButton = screen.getByText('Dashboard');
    fireEvent.click(dashboardButton);

    expect(mockOnDrawerToggle).toHaveBeenCalledTimes(1);
  });

  it('does not call onDrawerToggle when menu item is clicked on desktop', () => {
    // Mock desktop view
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1200,
    });

    render(<Sidebar {...defaultProps} />);

    const dashboardButton = screen.getByText('Dashboard');
    fireEvent.click(dashboardButton);

    expect(mockOnDrawerToggle).not.toHaveBeenCalled();
  });

  it('navigates to correct path when menu item is clicked', () => {
    render(<Sidebar {...defaultProps} />);

    const usersButton = screen.getByText('Users');
    fireEvent.click(usersButton);

    expect(window.location.pathname).toBe('/users');
  });

  it('renders mobile drawer when mobileOpen is true', () => {
    render(<Sidebar {...defaultProps} mobileOpen={true} />);

    const drawer = screen.getByRole('presentation');
    expect(drawer).toBeInTheDocument();
  });

  it('applies correct styles to active menu item', () => {
    mockUseLocation.mockReturnValue({ pathname: '/users' });
    
    render(<Sidebar {...defaultProps} />);

    const usersButton = screen.getByText('Users').closest('button');
    expect(usersButton).toHaveStyle({
      backgroundColor: expect.stringContaining('rgb'),
    });
  });
});