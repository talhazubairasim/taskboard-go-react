import { Outlet } from 'react-router-dom';
import { Header } from './Header';
import { Sidebar } from './Sidebar';

export const Layout = () => (
  <div className="flex min-h-screen flex-col">
    <Header />
    <div className="flex flex-1 flex-col sm:flex-row">
      <Sidebar />
      <main className="flex-1 overflow-x-auto p-4 sm:p-6">
        <Outlet />
      </main>
    </div>
  </div>
);
