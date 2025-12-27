export default function Header() {
  return (
    <header className="flex items-center justify-between px-8 py-4 bg-gray-900 border-b border-gray-800 shadow-sm">
      <h1 className="text-2xl font-semibold text-white tracking-tight">Distributed Job Scheduler</h1>
      <div className="flex items-center gap-4">
        <span className="text-sm text-gray-200 font-mono bg-gray-800 px-2 py-1">v1.0</span>
      </div>
    </header>
  );
}
