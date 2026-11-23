import type { Priority } from '../../types';

const colors: Record<Priority, string> = {
  URGENT: 'bg-red-100 text-red-800',
  HIGH: 'bg-orange-100 text-orange-800',
  MEDIUM: 'bg-yellow-100 text-yellow-800',
  LOW: 'bg-green-100 text-green-800',
};

export const PriorityBadge = ({ priority }: { priority: Priority }) => (
  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${colors[priority]}`}>
    {priority.charAt(0) + priority.slice(1).toLowerCase()}
  </span>
);
