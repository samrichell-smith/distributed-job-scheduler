interface StatCardProps {
  title: string;
  value: string | number;
  icon?: React.ReactNode;
  color?: string;
}

export default function StatCard({ title, value, icon, color = "bg-blue-100 text-blue-800" }: StatCardProps) {
  return (
    <div className={`flex items-center gap-4 p-6 rounded-lg shadow-sm border border-gray-200 ${color} bg-opacity-60`}> 
      <div className="text-3xl">{icon}</div>
      <div>
        <div className="text-sm font-medium uppercase tracking-wide text-gray-600">{title}</div>
        <div className="text-2xl font-bold">{value}</div>
      </div>
    </div>
  );
}
