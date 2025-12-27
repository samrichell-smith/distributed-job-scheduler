import clsx from "clsx";

interface StatCardProps {
  title: string;
  value: string | number;
  icon?: React.ReactNode;
  color?: string;
}

export default function StatCard({ title, value, icon }: StatCardProps) {
  return (
    <div className={clsx("flex items-center gap-4 p-6 shadow-sm border border-gray-200 bg-white min-h-[110px]")}> 
      <div className="text-3xl flex-shrink-0 text-gray-700">{icon}</div>
      <div>
        <div className="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-1">
          {title}
        </div>
        <div className="text-2xl font-bold text-gray-900 leading-tight">
          {value}
        </div>
      </div>
    </div>
  );
}
