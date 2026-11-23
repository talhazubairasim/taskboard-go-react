import { useState } from 'react';
import type { Task, User } from '../../types';
import { Avatar } from '../common/Avatar';
import { Modal } from '../common/Modal';
import { getErrorMessage } from '../../utils/errors';

interface TaskAssignmentProps {
  task: Task;
  users: User[];
  onAssign: (userId: string) => Promise<unknown>;
  onUnassign: () => Promise<unknown>;
  onClose: () => void;
}

/** Dialog for choosing who a task is assigned to. */
export const TaskAssignment = ({ task, users, onAssign, onUnassign, onClose }: TaskAssignmentProps) => {
  const [search, setSearch] = useState('');
  const [error, setError] = useState('');
  const q = search.trim().toLowerCase();
  const matches = users.filter((u) => !q || u.name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q));

  const run = async (action: () => Promise<unknown>) => {
    try {
      await action();
      onClose();
    } catch (err) {
      setError(getErrorMessage(err));
    }
  };

  return (
    <Modal title={`Assign "${task.title}"`} onClose={onClose}>
      {error && <div role="alert" className="mb-3 rounded-md bg-red-50 p-3 text-sm text-red-800">{error}</div>}
      <input aria-label="Search people" placeholder="Search by name or email" value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="mb-3 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
      <ul className="max-h-72 divide-y overflow-y-auto">
        {matches.map((u) => {
          const current = task.assignedTo?.id === u.id;
          return (
            <li key={u.id}>
              <button disabled={current} onClick={() => run(() => onAssign(u.id))}
                className="flex w-full items-center gap-3 px-2 py-2 text-left hover:bg-gray-50 disabled:cursor-default disabled:bg-indigo-50">
                <Avatar name={u.name} src={u.avatar} size="sm" />
                <span className="flex-1">
                  <span className="block text-sm font-medium text-gray-900">{u.name}</span>
                  <span className="block text-xs text-gray-500">{u.email}</span>
                </span>
                {current && <span className="text-xs font-medium text-indigo-700">Assigned</span>}
              </button>
            </li>
          );
        })}
        {matches.length === 0 && <li className="py-4 text-center text-sm text-gray-500">No one matches "{search}"</li>}
      </ul>
      {task.assignedTo && (
        <button onClick={() => run(onUnassign)} className="mt-4 text-sm font-medium text-red-600 hover:text-red-700">
          Remove assignee
        </button>
      )}
    </Modal>
  );
};
