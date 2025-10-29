import { useState } from 'react';
import { useQuery } from '@apollo/client';
import { GET_TASKS, GET_USERS } from '../graphql/queries';
import { TaskList } from '../components/tasks/TaskList';
import { TaskForm } from '../components/tasks/TaskForm';
import { Header } from '../components/layout/Header';
import { useAuth } from '../contexts/AuthContext';

export const Dashboard = () => {
  const { user } = useAuth();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string | null>(null);

  const { data: tasksData, loading: tasksLoading, refetch: refetchTasks } = useQuery(GET_TASKS, {
    variables: {
      filter: statusFilter ? { status: statusFilter } : undefined,
    },
  });

  const { data: usersData } = useQuery(GET_USERS);

  const handleTaskCreated = () => {
    setShowCreateForm(false);
    refetchTasks();
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="mb-6">
            <h1 className="text-3xl font-bold text-gray-900">
              Welcome back, {user?.name}!
            </h1>
            <p className="mt-1 text-sm text-gray-600">
              Manage your tasks and collaborate with your team
            </p>
          </div>

          {/* Filters and Create Button */}
          <div className="mb-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="flex gap-2">
              <button
                onClick={() => setStatusFilter(null)}
                className={`px-4 py-2 rounded-md text-sm font-medium ${
                  statusFilter === null
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                }`}
              >
                All
              </button>
              <button
                onClick={() => setStatusFilter('TODO')}
                className={`px-4 py-2 rounded-md text-sm font-medium ${
                  statusFilter === 'TODO'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                }`}
              >
                To Do
              </button>
              <button
                onClick={() => setStatusFilter('IN_PROGRESS')}
                className={`px-4 py-2 rounded-md text-sm font-medium ${
                  statusFilter === 'IN_PROGRESS'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                }`}
              >
                In Progress
              </button>
              <button
                onClick={() => setStatusFilter('DONE')}
                className={`px-4 py-2 rounded-md text-sm font-medium ${
                  statusFilter === 'DONE'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                }`}
              >
                Done
              </button>
            </div>

            <button
              onClick={() => setShowCreateForm(true)}
              className="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm font-medium hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              + Create Task
            </button>
          </div>

          {/* Create Task Form Modal */}
          {showCreateForm && (
            <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50">
              <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-xl font-bold text-gray-900">Create New Task</h2>
                  <button
                    onClick={() => setShowCreateForm(false)}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                <TaskForm 
                  users={usersData?.users || []}
                  onComplete={handleTaskCreated}
                  onCancel={() => setShowCreateForm(false)}
                />
              </div>
            </div>
          )}

          {/* Tasks List */}
          {tasksLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
            </div>
          ) : (
            <TaskList 
              tasks={tasksData?.tasks || []} 
              users={usersData?.users || []}
              onTaskUpdated={refetchTasks}
            />
          )}
        </div>
      </main>
    </div>
  );
};