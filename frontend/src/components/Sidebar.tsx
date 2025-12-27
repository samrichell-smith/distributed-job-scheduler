import { Link, useLocation } from "react-router-dom";
import { FaTasks, FaListUl, FaChartBar } from "react-icons/fa";

export default function Sidebar() {
  const location = useLocation();
  return (
    <aside className="h-screen w-60 bg-gray-900 text-gray-100 flex flex-col shadow-inner border-r border-gray-800 relative">
      {/* Logo/Header */}
    <div className="flex items-center gap-3 px-6 py-7 border-b border-gray-800 bg-gray-900/90 z-10 relative tracking-tight select-none">
  <FaTasks className="text-gray-300" size={32} />
        <span className="text-gray-100 font-semibold">Job Scheduler</span>
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
              className={`flex items-center gap-3 px-3 py-2 rounded-none font-medium transition-colors hover:bg-gray-800/40 hover:text-gray-100 focus:bg-gray-800/60 focus:outline-none ${location.pathname === '/' ? 'bg-gray-800/30 text-gray-100' : 'text-gray-300'}`}
            >
              <FaListUl className="text-gray-300" />
              Jobs
            </Link>
          </li>
          <li>
            <Link
              to="/analytics"
              className={`flex items-center gap-3 px-3 py-2 rounded-none font-medium transition-colors hover:bg-gray-800/40 hover:text-gray-100 focus:bg-gray-800/60 focus:outline-none ${location.pathname === '/analytics' ? 'bg-gray-800/30 text-gray-100' : 'text-gray-300'}`}
            >
              <FaChartBar className="text-gray-300" />
              Analytics
            </Link>
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
