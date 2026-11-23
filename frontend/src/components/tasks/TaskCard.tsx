import { STATUSES, type Task, type TaskStatus } from '../../types';
import { Avatar } from '../common/Avatar';
import { PriorityBadge } from '../common/PriorityBadge';
import { formatDate, isOverdue } from '../../utils/format';

interface TaskCardProps {
  task: Task;
  currentUserId: string;
  onMove: (status: TaskStatus) => void;
  onEdit: () => void;
  onAssign: () => void;
  onDelete: () => void;
}

export const TaskCard = ({ task, currentUserId, onMove, onEdit, onAssign, onDelete }: TaskCardProps) => {
  const isCreator = task.createdBy.id === currentUserId;
  const canEdit = isCreator || task.assignedTo?.id === currentUserId;

  return (
    <article
      draggable={canEdit}
      onDragStart={(e) => e.dataTransfer.setData('text/plain', task.id)}
      aria-label={task.title}
      className={`rounded-lg border border-gray-200 bg-white p-3 shadow-sm ${canEdit ? 'cursor-grab' : ''}`}
    >
      <div className="mb-1 flex items-start justify-between gap-2">
        <h3 className="text-sm font-semibold text-gray-900">{task.title}</h3>
        <PriorityBadge priority={task.priority} />
      </div>
      {task.description && <p className="mb-2 line-clamp-3 text-xs text-gray-600">{task.description}</p>}
      <div className="flex items-center justify-between text-xs text-gray-500">
        {task.dueDate ? (
          <span className={isOverdue(task) ? 'font-medium text-red-600' : ''}>
            Due {formatDate(task.dueDate)}
          </span>
        ) : (
          <span />
        )}
        {task.assignedTo ? (
          <span className="flex items-center gap-1" title={`Assigned to ${task.assignedTo.name}`}>
            <Avatar name={task.assignedTo.name} src={task.assignedTo.avatar} size="sm" />
          </span>
        ) : (
          <span className="italic">Unassigned</span>
        )}
      </div>
      {canEdit && (
        <div className="mt-3 flex flex-wrap items-center gap-2 border-t border-gray-100 pt-2">
          <label className="sr-only" htmlFor={`move-${task.id}`}>Move {task.title}</label>
          <select id={`move-${task.id}`} value={task.status} onChange={(e) => onMove(e.target.value as TaskStatus)}
            className="rounded border border-gray-200 px-1 py-0.5 text-xs">
            {STATUSES.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
          </select>
          <button onClick={onEdit} className="text-xs font-medium text-indigo-600 hover:text-indigo-800">Edit</button>
          <button onClick={onAssign} className="text-xs font-medium text-indigo-600 hover:text-indigo-800">Assign</button>
          {isCreator && (
            <button onClick={onDelete} className="ml-auto text-xs font-medium text-red-600 hover:text-red-800">Delete</button>
          )}
        </div>
      )}
    </article>
  );
};
