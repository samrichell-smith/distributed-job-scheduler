import { Link, useLocation } from "react-router-dom";

export default function TopTabs() {
  const location = useLocation();
  return (
    <nav className="w-full border-b border-gray-200 bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex -mb-px space-x-8">
          <Link
            to="/"
            className={`inline-flex items-center py-3 px-1 border-b-2 text-sm font-medium ${location.pathname === '/' ? 'border-gray-800 text-gray-900' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}`}
          >
            Jobs
          </Link>
          <Link
            to="/analytics"
            className={`inline-flex items-center py-3 px-1 border-b-2 text-sm font-medium ${location.pathname === '/analytics' ? 'border-gray-800 text-gray-900' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}`}
          >
            Analytics
          </Link>
        </div>
      </div>
    </nav>
  );
}
