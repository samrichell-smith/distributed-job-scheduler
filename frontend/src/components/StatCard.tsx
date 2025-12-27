import clsx from "clsx";

interface StatCardProps {
  title: string;
  value: string | number;
  icon?: React.ReactNode;
  color?: string;
  large?: boolean;
}

export default function StatCard({ title, value, icon, large }: StatCardProps) {
  return (
    <div className={clsx("flex items-center gap-4 shadow-sm border border-gray-200 bg-white", large ? 'p-8 min-h-[150px]' : 'p-6 min-h-[110px]')}> 
      {icon && <div className={clsx(large ? 'text-4xl' : 'text-3xl', 'flex-shrink-0 text-gray-700')}>{icon}</div>}
      <div>
        <div className={clsx('text-xs font-semibold uppercase tracking-wide text-gray-500 mb-1', large && 'text-sm')}>
          {title}
        </div>
        <div className={clsx('text-2xl font-bold text-gray-900 leading-tight', large && 'text-4xl')}>
          {value}
        </div>
      </div>
    </div>
  );
}
