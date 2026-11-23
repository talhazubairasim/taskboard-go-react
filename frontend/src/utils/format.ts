import type { Task } from '../types';

/** Formats an ISO date as e.g. "Jan 2, 2030" (dates are stored at UTC midnight). */
export const formatDate = (iso: string) =>
  new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric', timeZone: 'UTC' });

/** A task is overdue when its due date has passed and it is not done. */
export const isOverdue = (task: Task, now = new Date()) =>
  !!task.dueDate && task.status !== 'DONE' && new Date(task.dueDate) < now;

/** Converts an <input type="date"> value to the RFC 3339 timestamp the API expects. */
export const toDueDate = (date: string) => (date ? `${date}T00:00:00Z` : null);
