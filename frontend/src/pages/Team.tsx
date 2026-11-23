import { useTasks } from '../hooks/useTasks';
import { useUsers } from '../hooks/useUsers';
import { UserList } from '../components/users/UserList';
import { Spinner } from '../components/common/Spinner';

export const Team = () => {
  const { users, loading } = useUsers();
  const { tasks } = useTasks();
  return (
    <div>
      <h1 className="mb-4 text-2xl font-bold text-gray-900">Team</h1>
      {loading && users.length === 0 ? <Spinner /> : <UserList users={users} tasks={tasks} />}
    </div>
  );
};
