import { useState } from 'react';
import { useMutation } from '@apollo/client';
import { UPDATE_TASK, DELETE_TASK, ASSIGN_TASK } from '../../graphql/mutations';

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

interface TaskItemProps {
  task: Task;
  users: User[];
  onTaskUpdated: () => void;
}

export const TaskItem = ({ task, users, onTaskUpdated }: TaskItemProps) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showAssignMenu, setShowAssignMenu] = useState(false);

  const [updateTask] = useMutation(UPDATE_TASK, {
    onCompleted: onTaskUpdated,
  });

  const [deleteTask] = useMutation(DELETE_TASK, {
    onCompleted: onTaskUpdated,
  });

  const [assignTask] = useMutation(ASSIGN_TASK, {
    onCompleted: onTaskUpdated,
  });

  const handleStatusChange = (newStatus: string) => {
    updateTask({
      variables: {
        id: task.id,
        input: { status: newStatus },
      },
    });
  };

  const handleDelete = () => {
    if (window.confirm('Are you sure you want to delete this task?')) {
      deleteTask({ variables: { id: task.id } });
    }
  };

  const handleAssign = (userId: string) => {
    assignTask({
      variables: {
        taskId: task.id,
        userId,
      },
    });
    setShowAssignMenu(false);
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'URGENT':
        return 'bg-red-100 text-red-800';
      case 'HIGH':
        return 'bg-orange-100 text-orange-800';
      case 'MEDIUM':
        return 'bg-yellow-100 text-yellow-800';
      case 'LOW':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'DONE':
        return 'bg-green-100 text-green-800';
      case 'IN_PROGRESS':
        return 'bg-blue-100 text-blue-800';
      case 'REVIEW':
        return 'bg-purple-100 text-purple-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
      <div className="p-4">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-2">
              <h3 className="text-lg font-semibold text-gray-900">{task.title}</h3>
              <span className={`px-2 py-1 text-xs font-medium rounded-full ${getPriorityColor(task.priority)}`}>
                {task.priority}
              </span>
              <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(task.status)}`}>
                {task.status.replace('_', ' ')}
              </span>
            </div>

            {task.description && (
              <p className="text-sm text-gray-600 mb-2">
                {isExpanded ? task.description : task.description.slice(0, 100)}
                {task.description.length > 100 && (
                  <button
                    onClick={() => setIsExpanded(!isExpanded)}
                    className="text-indigo-600 hover:text-indigo-800 ml-1"
                  >
                    {isExpanded ? 'Show less' : 'Show more'}
                  </button>
                )}
              </p>
            )}

            <div className="flex items-center gap-4 text-xs text-gray-500">
              <span>Created by {task.createdBy.name}</span>
              {task.assignedTo && <span>Assigned to {task.assignedTo.name}</span>}
              {task.dueDate && (
                <span>Due: {new Date(task.dueDate).toLocaleDateString()}</span>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2 ml-4">
            {/* Status Update Dropdown */}
            <select
              value={task.status}
              onChange={(e) => handleStatusChange(e.target.value)}
              className="text-sm border border-gray-300 rounded-md px-2 py-1 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="TODO">To Do</option>
              <option value="IN_PROGRESS">In Progress</option>
              <option value="REVIEW">Review</option>
              <option value="DONE">Done</option>
            </select>

            {/* Assign Button */}
            <div className="relative">
              <button
                onClick={() => setShowAssignMenu(!showAssignMenu)}
                className="p-2 text-gray-400 hover:text-gray-600"
                title="Assign task"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              </button>

              {showAssignMenu && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg z-10 border border-gray-200">
                  <div className="py-1">
                    {users.map((user) => (
                      <button
                        key={user.id}
                        onClick={() => handleAssign(user.id)}
                        className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                      >
                        {user.name}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {/* Delete Button */}
            <button
              onClick={handleDelete}
              className="p-2 text-red-400 hover:text-red-600"
              title="Delete task"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};