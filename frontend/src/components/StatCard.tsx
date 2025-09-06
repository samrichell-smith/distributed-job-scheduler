import clsx from "clsx";

interface StatCardProps {
  title: string;
  value: string | number;
  icon?: React.ReactNode;
  color?: string;
}

export default function StatCard({ title, value, icon, color = "from-blue-400 to-blue-600" }: StatCardProps) {
  return (
    <div
      className={clsx(
        "flex items-center gap-4 p-6 rounded-xl shadow-lg border border-gray-200 bg-gradient-to-br",
        color,
        "transition-transform hover:scale-105 hover:shadow-2xl cursor-pointer min-h-[110px]"
      )}
    >
      <div className="text-4xl flex-shrink-0 drop-shadow-lg opacity-90">{icon}</div>
      <div>
        <div className="text-xs font-semibold uppercase tracking-wide text-white/80 mb-1 drop-shadow">
          {title}
        </div>
        <div className="text-3xl font-extrabold text-white drop-shadow-lg leading-tight">
          {value}
        </div>
      </div>
    </div>
  );
}
