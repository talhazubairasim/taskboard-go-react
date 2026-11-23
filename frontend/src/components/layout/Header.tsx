import { useAuth } from '../../contexts/AuthContext';
import { Avatar } from '../common/Avatar';

export const Header = () => {
  const { user, logout } = useAuth();
  return (
    <header className="bg-white shadow-sm">
      <div className="flex items-center justify-between px-4 py-3 sm:px-6">
        <h1 className="text-2xl font-bold text-indigo-600">TaskBoard</h1>
        {user && (
          <div className="flex items-center gap-3">
            <Avatar name={user.name} src={user.avatar} />
            <span className="hidden text-sm font-medium text-gray-700 sm:inline">{user.name}</span>
            <button onClick={logout}
              className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50">
              Logout
            </button>
          </div>
        )}
      </div>
    </header>
  );
};
