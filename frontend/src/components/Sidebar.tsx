import { FaTasks, FaListUl } from "react-icons/fa";

export default function Sidebar() {
  return (
    <aside className="h-screen w-56 bg-gray-900 text-white flex flex-col shadow-lg">
      <div className="flex items-center gap-2 px-6 py-6 text-xl font-bold border-b border-gray-800">
        <FaTasks className="text-blue-400" />
        <span>Job Scheduler</span>
      </div>
      <nav className="flex-1 px-4 py-6">
        <ul className="space-y-4">
          <li>
            <a href="#" className="flex items-center gap-2 hover:text-blue-400 font-medium">
              <FaListUl className="text-blue-300" />
              Jobs
            </a>
          </li>
        </ul>
      </nav>
      <div className="px-6 py-4 text-xs text-gray-400 border-t border-gray-800">
        &copy; {new Date().getFullYear()} Scheduler UI
      </div>
    </aside>
  );
}
