export default function Header() {
  return (
    <header className="flex items-center justify-between px-8 py-4 bg-gradient-to-r from-blue-700 to-blue-500 border-b border-blue-700 shadow-md">
      <h1 className="text-2xl font-bold text-white tracking-tight drop-shadow">Distributed Job Scheduler</h1>
      <div className="flex items-center gap-4">
        <span className="text-sm text-blue-100 font-mono bg-blue-900/40 px-2 py-1 rounded">v1.0</span>
      </div>
    </header>
  );
}
