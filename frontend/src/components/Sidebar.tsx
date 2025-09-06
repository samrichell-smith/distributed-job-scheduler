import { Link, useLocation } from "react-router-dom";
import { FaTasks, FaListUl, FaChartBar, FaCog } from "react-icons/fa";

export default function Sidebar() {
  const location = useLocation();
  return (
    <aside className="h-screen w-60 bg-gradient-to-b from-gray-900 to-gray-800 text-gray-100 flex flex-col shadow-2xl border-r border-gray-800 relative">
      {/* Decorative background accent */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute -top-10 -left-10 w-40 h-40 bg-blue-900 opacity-20 rounded-full blur-2xl" />
        <div className="absolute bottom-0 right-0 w-32 h-32 bg-blue-700 opacity-10 rounded-full blur-2xl" />
      </div>
      {/* Logo/Header */}
    <div className="flex items-center gap-3 px-6 py-7 border-b border-gray-800 bg-gray-900/90 z-10 relative tracking-tight select-none">
  <FaTasks className="text-blue-500 drop-shadow-md" size={32} />
        <span className="bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent">
          Job Scheduler
        </span>
      </div>
      {/* Navigation */}
      <nav className="flex-1 px-4 py-8 z-10 relative">
        <div className="mb-4 text-xs uppercase tracking-widest text-gray-400 font-semibold pl-2">
          Main
        </div>
        <ul className="space-y-2">
          <li>
            <Link
              to="/"
              className={`flex items-center gap-3 px-3 py-2 rounded-lg font-medium transition-all hover:bg-blue-900/20 hover:text-blue-400 focus:bg-blue-900/30 focus:outline-none ${location.pathname === '/' ? 'bg-blue-900/10 text-blue-400' : ''}`}
            >
              <FaListUl className="text-blue-400" />
              Jobs
            </Link>
          </li>
          <li>
            <Link
              to="/analytics"
              className={`flex items-center gap-3 px-3 py-2 rounded-lg font-medium transition-all hover:bg-blue-900/20 hover:text-blue-400 focus:bg-blue-900/30 focus:outline-none ${location.pathname === '/analytics' ? 'bg-blue-900/10 text-blue-400' : ''}`}
            >
              <FaChartBar className="text-cyan-400" />
              Analytics
            </Link>
          </li>
        </ul>
        <div className="mt-8 text-xs uppercase tracking-widest text-gray-400 font-semibold pl-2">
          Settings
        </div>
        <ul className="space-y-2 mt-2">
          <li>
            <a
              href="#"
              className="flex items-center gap-3 px-3 py-2 rounded-lg font-medium transition-all hover:bg-blue-900/20 hover:text-blue-400 focus:bg-blue-900/30 focus:outline-none"
            >
              <FaCog className="text-gray-400" />
              Preferences
            </a>
          </li>
        </ul>
      </nav>
      {/* Footer */}
      <div className="px-6 py-4 text-xs text-gray-400 border-t border-gray-800 bg-gray-900/90 z-10 relative select-none">
        &copy; {new Date().getFullYear()} Scheduler UI
      </div>
    </aside>
  );
}
