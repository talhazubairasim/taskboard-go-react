import { useState, type FormEvent } from 'react';
import { PRIORITIES, STATUSES, type Priority, type Task, type TaskStatus, type User } from '../../types';
import type { TaskInput } from '../../hooks/useTasks';
import { getErrorMessage } from '../../utils/errors';
import { toDueDate } from '../../utils/format';

interface TaskFormProps {
  users: User[];
  /** When set, the form edits this task instead of creating one. */
  task?: Task;
  onSubmit: (input: TaskInput) => Promise<unknown>;
  onDone: () => void;
  onCancel: () => void;
}

const inputClass =
  'block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500';

export const TaskForm = ({ users, task, onSubmit, onDone, onCancel }: TaskFormProps) => {
  const [title, setTitle] = useState(task?.title ?? '');
  const [description, setDescription] = useState(task?.description ?? '');
  const [status, setStatus] = useState<TaskStatus>(task?.status ?? 'TODO');
  const [priority, setPriority] = useState<Priority>(task?.priority ?? 'MEDIUM');
  const [assignedToId, setAssignedToId] = useState(task?.assignedTo?.id ?? '');
  const [dueDate, setDueDate] = useState(task?.dueDate ? task.dueDate.slice(0, 10) : '');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError('Title is required');
      return;
    }
    setError('');
    setSaving(true);
    try {
      await onSubmit({
        title: title.trim(),
        description: description.trim() || (task ? '' : null),
        status,
        priority,
        // An empty string tells the API to clear the assignee when editing.
        assignedToId: assignedToId || (task ? '' : null),
        dueDate: toDueDate(dueDate),
      });
      onDone();
    } catch (err) {
      setError(getErrorMessage(err));
    } finally {
      setSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4" noValidate>
      {error && (
        <div role="alert" className="rounded-md bg-red-50 p-3 text-sm text-red-800">
          {error}
        </div>
      )}
      <div>
        <label htmlFor="task-title" className="mb-1 block text-sm font-medium text-gray-700">Title</label>
        <input id="task-title" className={inputClass} value={title} onChange={(e) => setTitle(e.target.value)} maxLength={500} />
      </div>
      <div>
        <label htmlFor="task-description" className="mb-1 block text-sm font-medium text-gray-700">Description</label>
        <textarea id="task-description" rows={3} className={inputClass} value={description}
          onChange={(e) => setDescription(e.target.value)} />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label htmlFor="task-status" className="mb-1 block text-sm font-medium text-gray-700">Status</label>
          <select id="task-status" className={inputClass} value={status} onChange={(e) => setStatus(e.target.value as TaskStatus)}>
            {STATUSES.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
          </select>
        </div>
        <div>
          <label htmlFor="task-priority" className="mb-1 block text-sm font-medium text-gray-700">Priority</label>
          <select id="task-priority" className={inputClass} value={priority} onChange={(e) => setPriority(e.target.value as Priority)}>
            {PRIORITIES.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
          </select>
        </div>
        <div>
          <label htmlFor="task-assignee" className="mb-1 block text-sm font-medium text-gray-700">Assignee</label>
          <select id="task-assignee" className={inputClass} value={assignedToId} onChange={(e) => setAssignedToId(e.target.value)}>
            <option value="">Unassigned</option>
            {users.map((u) => <option key={u.id} value={u.id}>{u.name}</option>)}
          </select>
        </div>
        <div>
          <label htmlFor="task-due" className="mb-1 block text-sm font-medium text-gray-700">Due date</label>
          <input id="task-due" type="date" className={inputClass} value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
        </div>
      </div>
      <div className="flex justify-end gap-2 pt-2">
        <button type="button" onClick={onCancel} className="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
          Cancel
        </button>
        <button type="submit" disabled={saving}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50">
          {saving ? 'Saving...' : task ? 'Save changes' : 'Create task'}
        </button>
      </div>
    </form>
  );
};
