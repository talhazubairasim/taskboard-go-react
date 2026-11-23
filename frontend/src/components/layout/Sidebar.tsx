import { NavLink } from 'react-router-dom';

const items = [
  { to: '/dashboard', label: 'Board' },
  { to: '/team', label: 'Team' },
  { to: '/profile', label: 'Profile' },
];

export const Sidebar = () => (
  <nav aria-label="Main" className="border-b border-gray-200 bg-white sm:w-48 sm:border-b-0 sm:border-r">
    <ul className="flex gap-1 p-2 sm:flex-col">
      {items.map((item) => (
        <li key={item.to}>
          <NavLink
            to={item.to}
            className={({ isActive }) =>
              `block rounded-md px-3 py-2 text-sm font-medium ${
                isActive ? 'bg-indigo-50 text-indigo-700' : 'text-gray-700 hover:bg-gray-100'
              }`
            }
          >
            {item.label}
          </NavLink>
        </li>
      ))}
    </ul>
  </nav>
);
