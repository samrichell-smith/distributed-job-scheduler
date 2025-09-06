export default function Header() {
  return (
    <header className="flex items-center justify-between px-8 py-4 bg-white border-b border-gray-200 shadow-sm">
      <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Distributed Job Scheduler</h1>
      <div className="flex items-center gap-4">
        <span className="text-sm text-gray-500">v1.0</span>
      </div>
    </header>
  );
}
