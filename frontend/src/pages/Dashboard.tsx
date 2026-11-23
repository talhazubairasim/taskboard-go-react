import { useMemo, useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useTaskActions, useTasks } from '../hooks/useTasks';
import { useRealtime } from '../hooks/useRealtime';
import { useUsers } from '../hooks/useUsers';
import { PRIORITIES, type Priority, type Task } from '../types';
import { KanbanBoard } from '../components/tasks/KanbanBoard';
import { TaskForm } from '../components/tasks/TaskForm';
import { TaskAssignment } from '../components/tasks/TaskAssignment';
import { Modal } from '../components/common/Modal';
import { Spinner } from '../components/common/Spinner';
import { getErrorMessage } from '../utils/errors';

type Dialog = { kind: 'create' } | { kind: 'edit'; task: Task } | { kind: 'assign'; task: Task } | null;

export const Dashboard = () => {
  const { user } = useAuth();
  const { tasks, loading, error } = useTasks();
  const { users } = useUsers();
  const actions = useTaskActions();
  const { connected } = useRealtime();

  const [dialog, setDialog] = useState<Dialog>(null);
  const [search, setSearch] = useState('');
  const [priority, setPriority] = useState<Priority | ''>('');
  const [mineOnly, setMineOnly] = useState(false);
  const [actionError, setActionError] = useState('');

  const visible = useMemo(() => {
    const q = search.trim().toLowerCase();
    return tasks.filter(
      (t) =>
        (!q || t.title.toLowerCase().includes(q) || (t.description ?? '').toLowerCase().includes(q)) &&
        (!priority || t.priority === priority) &&
        (!mineOnly || t.assignedTo?.id === user?.id || t.createdBy.id === user?.id),
    );
  }, [tasks, search, priority, mineOnly, user]);

  if (!user) return null;
  const close = () => setDialog(null);
  const attempt = (p: Promise<unknown>) => p.catch((err) => setActionError(getErrorMessage(err)));

  return (
    <div>
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Welcome back, {user.name}!</h1>
          <p className="mt-1 flex items-center gap-2 text-sm text-gray-600">
            <span className={`inline-block h-2 w-2 rounded-full ${connected ? 'bg-green-500' : 'bg-gray-400'}`} />
            {connected ? 'Live updates on' : 'Live updates paused'}
          </p>
        </div>
        <button onClick={() => setDialog({ kind: 'create' })}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700">
          + Create Task
        </button>
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-3">
        <input aria-label="Search tasks" placeholder="Search tasks" value={search} onChange={(e) => setSearch(e.target.value)}
          className="w-56 rounded-md border border-gray-300 px-3 py-1.5 text-sm" />
        <select aria-label="Filter by priority" value={priority} onChange={(e) => setPriority(e.target.value as Priority | '')}
          className="rounded-md border border-gray-300 px-2 py-1.5 text-sm">
          <option value="">All priorities</option>
          {PRIORITIES.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
        </select>
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <input type="checkbox" checked={mineOnly} onChange={(e) => setMineOnly(e.target.checked)} />
          Only my tasks
        </label>
      </div>

      {(error || actionError) && (
        <div role="alert" className="mb-4 flex justify-between rounded-md bg-red-50 p-3 text-sm text-red-800">
          {actionError || getErrorMessage(error)}
          {actionError && <button onClick={() => setActionError('')} aria-label="Dismiss">×</button>}
        </div>
      )}

      {loading && tasks.length === 0 ? (
        <Spinner />
      ) : (
        <KanbanBoard
          tasks={visible}
          currentUserId={user.id}
          onMove={(task, status) => attempt(actions.moveTask(task, status))}
          onEdit={(task) => setDialog({ kind: 'edit', task })}
          onAssign={(task) => setDialog({ kind: 'assign', task })}
          onDelete={(task) => {
            if (window.confirm(`Delete "${task.title}"? This cannot be undone.`)) attempt(actions.deleteTask(task.id));
          }}
        />
      )}

      {dialog?.kind === 'create' && (
        <Modal title="Create New Task" onClose={close}>
          <TaskForm users={users} onSubmit={actions.createTask} onDone={close} onCancel={close} />
        </Modal>
      )}
      {dialog?.kind === 'edit' && (
        <Modal title="Edit Task" onClose={close}>
          <TaskForm users={users} task={dialog.task} onSubmit={(input) => actions.updateTask(dialog.task.id, input)}
            onDone={close} onCancel={close} />
        </Modal>
      )}
      {dialog?.kind === 'assign' && (
        <TaskAssignment task={dialog.task} users={users} onClose={close}
          onAssign={(userId) => actions.assignTask(dialog.task.id, userId)}
          onUnassign={() => actions.unassignTask(dialog.task.id)} />
      )}
    </div>
  );
};
