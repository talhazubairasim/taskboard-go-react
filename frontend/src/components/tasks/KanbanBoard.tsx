import { useState } from 'react';
import { STATUSES, type Task, type TaskStatus } from '../../types';
import { TaskCard } from './TaskCard';

interface KanbanBoardProps {
  tasks: Task[];
  currentUserId: string;
  onMove: (task: Task, status: TaskStatus) => void;
  onEdit: (task: Task) => void;
  onAssign: (task: Task) => void;
  onDelete: (task: Task) => void;
}

/** Four status columns. Cards can be dragged between columns or moved with their status menu. */
export const KanbanBoard = ({ tasks, currentUserId, onMove, onEdit, onAssign, onDelete }: KanbanBoardProps) => {
  const [over, setOver] = useState<TaskStatus | null>(null);

  const handleDrop = (status: TaskStatus, taskId: string) => {
    setOver(null);
    const task = tasks.find((t) => t.id === taskId);
    if (task && task.status !== status) onMove(task, status);
  };

  return (
    <div className="grid min-w-[56rem] grid-cols-4 gap-4">
      {STATUSES.map(({ value, label }) => {
        const column = tasks.filter((t) => t.status === value);
        return (
          <section
            key={value}
            aria-label={label}
            onDragOver={(e) => {
              e.preventDefault();
              setOver(value);
            }}
            onDragLeave={() => setOver(null)}
            onDrop={(e) => {
              e.preventDefault();
              handleDrop(value, e.dataTransfer.getData('text/plain'));
            }}
            className={`flex min-h-[20rem] flex-col rounded-lg p-3 ${over === value ? 'bg-indigo-100' : 'bg-gray-100'}`}
          >
            <h2 className="mb-3 flex items-center justify-between text-sm font-semibold text-gray-700">
              {label}
              <span className="rounded-full bg-white px-2 py-0.5 text-xs text-gray-500">{column.length}</span>
            </h2>
            <div className="flex flex-col gap-3">
              {column.map((task) => (
                <TaskCard
                  key={task.id}
                  task={task}
                  currentUserId={currentUserId}
                  onMove={(status) => onMove(task, status)}
                  onEdit={() => onEdit(task)}
                  onAssign={() => onAssign(task)}
                  onDelete={() => onDelete(task)}
                />
              ))}
            </div>
          </section>
        );
      })}
    </div>
  );
};
