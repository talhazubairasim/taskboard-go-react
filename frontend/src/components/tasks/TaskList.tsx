import { TaskItem } from './TaskItem';

interface Task {
  id: string;
  title: string;
  description?: string;
  status: string;
  priority: string;
  createdBy: {
    id: string;
    name: string;
  };
  assignedTo?: {
    id: string;
    name: string;
  };
  dueDate?: string;
  createdAt: string;
}

interface User {
  id: string;
  name: string;
  email: string;
}

interface TaskListProps {
  tasks: Task[];
  users: User[];
  onTaskUpdated: () => void;
}

export const TaskList = ({ tasks, users, onTaskUpdated }: TaskListProps) => {
  if (tasks.length === 0) {
    return (
      <div className="text-center py-12">
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
          />
        </svg>
        <h3 className="mt-2 text-sm font-medium text-gray-900">No tasks</h3>
        <p className="mt-1 text-sm text-gray-500">
          Get started by creating a new task.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {tasks.map((task) => (
        <TaskItem 
          key={task.id} 
          task={task} 
          users={users}
          onTaskUpdated={onTaskUpdated}
        />
      ))}
    </div>
  );
};