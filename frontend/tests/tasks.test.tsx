import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ApolloError } from '@apollo/client';
import { TaskCard } from '../src/components/tasks/TaskCard';
import { KanbanBoard } from '../src/components/tasks/KanbanBoard';
import { TaskForm } from '../src/components/tasks/TaskForm';
import { TaskAssignment } from '../src/components/tasks/TaskAssignment';
import { alice, bob, makeTask } from './fixtures';

const noop = () => {};

describe('TaskCard', () => {
  it('shows details and all actions to the creator', async () => {
    const user = userEvent.setup();
    const onMove = vi.fn(), onEdit = vi.fn(), onAssign = vi.fn(), onDelete = vi.fn();
    render(<TaskCard task={makeTask({ dueDate: '2000-01-01T00:00:00Z', assignedTo: bob })} currentUserId={alice.id}
      onMove={onMove} onEdit={onEdit} onAssign={onAssign} onDelete={onDelete} />);
    expect(screen.getByText('Write tests')).toBeInTheDocument();
    expect(screen.getByText('Cover the board')).toBeInTheDocument();
    expect(screen.getByText(/Due/)).toHaveClass('text-red-600');
    expect(screen.getByTitle('Assigned to Bob')).toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText('Move Write tests'), 'DONE');
    expect(onMove).toHaveBeenCalledWith('DONE');
    await user.click(screen.getByText('Edit'));
    await user.click(screen.getByText('Assign'));
    await user.click(screen.getByText('Delete'));
    expect(onEdit).toHaveBeenCalled();
    expect(onAssign).toHaveBeenCalled();
    expect(onDelete).toHaveBeenCalled();
  });

  it('lets the assignee edit but not delete', () => {
    render(<TaskCard task={makeTask({ assignedTo: bob, dueDate: '2999-01-01T00:00:00Z', description: null })}
      currentUserId={bob.id} onMove={noop} onEdit={noop} onAssign={noop} onDelete={noop} />);
    expect(screen.getByText('Edit')).toBeInTheDocument();
    expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    expect(screen.getByText(/Due/)).not.toHaveClass('text-red-600');
  });

  it('is read-only for everyone else', () => {
    render(<TaskCard task={makeTask()} currentUserId="stranger" onMove={noop} onEdit={noop} onAssign={noop} onDelete={noop} />);
    expect(screen.queryByText('Edit')).not.toBeInTheDocument();
    expect(screen.getByText('Unassigned')).toBeInTheDocument();
    expect(screen.getByRole('article')).toHaveAttribute('draggable', 'false');
  });
});

describe('KanbanBoard', () => {
  const tasks = [makeTask({ id: 'a', title: 'A' }), makeTask({ id: 'b', title: 'B', status: 'DONE' })];

  it('groups tasks into columns', () => {
    render(<KanbanBoard tasks={tasks} currentUserId={alice.id} onMove={noop} onEdit={noop} onAssign={noop} onDelete={noop} />);
    expect(screen.getByRole('region', { name: 'To Do' })).toHaveTextContent('A');
    expect(screen.getByRole('region', { name: 'Done' })).toHaveTextContent('B');
    expect(screen.getByRole('region', { name: 'Review' })).toHaveTextContent('0');
  });

  it('moves a card when it is dropped on another column', () => {
    const onMove = vi.fn();
    render(<KanbanBoard tasks={tasks} currentUserId={alice.id} onMove={onMove} onEdit={noop} onAssign={noop} onDelete={noop} />);
    const store: Record<string, string> = {};
    const dataTransfer = { setData: (k: string, v: string) => (store[k] = v), getData: (k: string) => store[k] };
    fireEvent.dragStart(screen.getByRole('article', { name: 'A' }), { dataTransfer });
    const review = screen.getByRole('region', { name: 'Review' });
    fireEvent.dragOver(review, { dataTransfer });
    expect(review).toHaveClass('bg-indigo-100');
    fireEvent.dragLeave(review);
    expect(review).not.toHaveClass('bg-indigo-100');
    fireEvent.drop(review, { dataTransfer });
    expect(onMove).toHaveBeenCalledWith(tasks[0], 'REVIEW');

    onMove.mockClear();
    fireEvent.drop(screen.getByRole('region', { name: 'To Do' }), { dataTransfer }); // same column
    fireEvent.drop(review, { dataTransfer: { getData: () => 'unknown' } });
    expect(onMove).not.toHaveBeenCalled();
  });

  it('forwards card actions', async () => {
    const user = userEvent.setup();
    const onMove = vi.fn(), onEdit = vi.fn(), onAssign = vi.fn(), onDelete = vi.fn();
    render(<KanbanBoard tasks={[tasks[0]]} currentUserId={alice.id} onMove={onMove} onEdit={onEdit} onAssign={onAssign} onDelete={onDelete} />);
    await user.selectOptions(screen.getByLabelText('Move A'), 'IN_PROGRESS');
    await user.click(screen.getByText('Edit'));
    await user.click(screen.getByText('Assign'));
    await user.click(screen.getByText('Delete'));
    expect(onMove).toHaveBeenCalledWith(tasks[0], 'IN_PROGRESS');
    expect(onEdit).toHaveBeenCalledWith(tasks[0]);
    expect(onAssign).toHaveBeenCalledWith(tasks[0]);
    expect(onDelete).toHaveBeenCalledWith(tasks[0]);
  });
});

describe('TaskForm', () => {
  it('creates a task with every field', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue({});
    const onDone = vi.fn();
    render(<TaskForm users={[alice, bob]} onSubmit={onSubmit} onDone={onDone} onCancel={noop} />);
    await user.type(screen.getByLabelText('Title'), '  New task ');
    await user.type(screen.getByLabelText('Description'), 'Details');
    await user.selectOptions(screen.getByLabelText('Status'), 'REVIEW');
    await user.selectOptions(screen.getByLabelText('Priority'), 'URGENT');
    await user.selectOptions(screen.getByLabelText('Assignee'), bob.id);
    fireEvent.change(screen.getByLabelText('Due date'), { target: { value: '2030-02-03' } });
    await user.click(screen.getByRole('button', { name: 'Create task' }));
    await waitFor(() => expect(onDone).toHaveBeenCalled());
    expect(onSubmit).toHaveBeenCalledWith({
      title: 'New task', description: 'Details', status: 'REVIEW', priority: 'URGENT',
      assignedToId: bob.id, dueDate: '2030-02-03T00:00:00Z',
    });
  });

  it('sends nulls for empty optional fields on create', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue({});
    render(<TaskForm users={[]} onSubmit={onSubmit} onDone={noop} onCancel={noop} />);
    await user.type(screen.getByLabelText('Title'), 'Minimal');
    await user.click(screen.getByRole('button', { name: 'Create task' }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ description: null, assignedToId: null, dueDate: null })));
  });

  it('edits an existing task and clears fields with empty strings', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue({});
    render(<TaskForm users={[alice, bob]} task={makeTask({ assignedTo: bob, dueDate: '2030-05-06T00:00:00Z' })}
      onSubmit={onSubmit} onDone={noop} onCancel={noop} />);
    expect(screen.getByLabelText('Title')).toHaveValue('Write tests');
    expect(screen.getByLabelText('Due date')).toHaveValue('2030-05-06');
    await user.clear(screen.getByLabelText('Description'));
    await user.selectOptions(screen.getByLabelText('Assignee'), '');
    await user.click(screen.getByRole('button', { name: 'Save changes' }));
    await waitFor(() => expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ description: '', assignedToId: '', dueDate: '2030-05-06T00:00:00Z' })));
  });

  it('validates the title and reports server errors', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockRejectedValue(new ApolloError({ errorMessage: 'title must be between 1 and 500 characters' }));
    const onCancel = vi.fn();
    render(<TaskForm users={[]} onSubmit={onSubmit} onDone={noop} onCancel={onCancel} />);
    await user.click(screen.getByRole('button', { name: 'Create task' }));
    expect(screen.getByRole('alert')).toHaveTextContent('Title is required');
    await user.type(screen.getByLabelText('Title'), 'x');
    await user.click(screen.getByRole('button', { name: 'Create task' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('title must be between');
    await user.click(screen.getByText('Cancel'));
    expect(onCancel).toHaveBeenCalled();
  });
});

describe('TaskAssignment', () => {
  it('searches people and assigns one', async () => {
    const user = userEvent.setup();
    const onAssign = vi.fn().mockResolvedValue({});
    const onClose = vi.fn();
    render(<TaskAssignment task={makeTask()} users={[alice, bob]} onAssign={onAssign} onUnassign={vi.fn()} onClose={onClose} />);
    await user.type(screen.getByLabelText('Search people'), 'bob@');
    expect(screen.queryByText('Alice')).not.toBeInTheDocument();
    await user.click(screen.getByText('Bob'));
    expect(onAssign).toHaveBeenCalledWith(bob.id);
    await waitFor(() => expect(onClose).toHaveBeenCalled());
    expect(screen.queryByText('Remove assignee')).not.toBeInTheDocument();
  });

  it('marks the current assignee and can remove them', async () => {
    const user = userEvent.setup();
    const onUnassign = vi.fn().mockResolvedValue({});
    const onClose = vi.fn();
    render(<TaskAssignment task={makeTask({ assignedTo: bob })} users={[alice, bob]} onAssign={vi.fn()} onUnassign={onUnassign} onClose={onClose} />);
    expect(screen.getByText('Assigned')).toBeInTheDocument();
    await user.click(screen.getByText('Remove assignee'));
    expect(onUnassign).toHaveBeenCalled();
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });

  it('shows an empty state and errors', async () => {
    const user = userEvent.setup();
    const onAssign = vi.fn().mockRejectedValue(new Error('assignee does not exist'));
    render(<TaskAssignment task={makeTask()} users={[alice]} onAssign={onAssign} onUnassign={vi.fn()} onClose={noop} />);
    await user.click(screen.getByText('Alice'));
    expect(await screen.findByRole('alert')).toHaveTextContent('assignee does not exist');
    await user.type(screen.getByLabelText('Search people'), 'zzz');
    expect(screen.getByText(/No one matches/)).toBeInTheDocument();
  });
});
