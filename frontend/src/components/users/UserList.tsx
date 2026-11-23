import { useMemo, useState } from 'react';
import type { Task, User } from '../../types';
import { Avatar } from '../common/Avatar';

interface UserListProps {
  users: User[];
  tasks: Task[];
}

/** The team: everyone on the board with how many open tasks they hold. */
export const UserList = ({ users, tasks }: UserListProps) => {
  const [search, setSearch] = useState('');

  const counts = useMemo(() => {
    const open: Record<string, number> = {};
    const done: Record<string, number> = {};
    for (const t of tasks) {
      if (!t.assignedTo) continue;
      const bucket = t.status === 'DONE' ? done : open;
      bucket[t.assignedTo.id] = (bucket[t.assignedTo.id] ?? 0) + 1;
    }
    return { open, done };
  }, [tasks]);

  const q = search.trim().toLowerCase();
  const visible = users.filter((u) => !q || u.name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q));

  return (
    <div>
      <input aria-label="Search team" placeholder="Search by name or email" value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="mb-4 block w-full max-w-sm rounded-md border border-gray-300 px-3 py-2 text-sm" />
      <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wide text-gray-500">
            <tr>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Email</th>
              <th className="px-4 py-3 text-right">Open tasks</th>
              <th className="px-4 py-3 text-right">Done</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {visible.map((u) => (
              <tr key={u.id}>
                <td className="flex items-center gap-2 px-4 py-3 font-medium text-gray-900">
                  <Avatar name={u.name} src={u.avatar} size="sm" />
                  {u.name}
                </td>
                <td className="px-4 py-3 text-gray-600">{u.email}</td>
                <td className="px-4 py-3 text-right">{counts.open[u.id] ?? 0}</td>
                <td className="px-4 py-3 text-right">{counts.done[u.id] ?? 0}</td>
              </tr>
            ))}
            {visible.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-gray-500">No team members found</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
