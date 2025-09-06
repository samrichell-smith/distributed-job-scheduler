import { useState, useEffect } from 'react';
import { type JobStats, fetchJobStats } from '../services/api';
import StatCard from './StatCard';
import { BiTask, BiCheckCircle, BiError, BiTime } from 'react-icons/bi';

export default function Stats() {
  const [stats, setStats] = useState<JobStats>({
    completed: 0,
    pending: 0,
    failed: 0,
  });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadStats = async () => {
      try {
        const data = await fetchJobStats();
        setStats(data);
      } catch (err) {
        console.error('Failed to fetch stats:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadStats();
    // Update stats every 5 seconds
    const interval = setInterval(loadStats, 5000);
    return () => clearInterval(interval);
  }, []);

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 p-4">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="animate-pulse">
            <div className="bg-gray-200 rounded-xl h-[110px]" />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 p-4">
      <StatCard
        title="Total Jobs"
        value={stats.completed + stats.pending + stats.failed}
        icon={<BiTask />}
        color="from-blue-400 to-blue-600"
      />
      <StatCard
        title="Completed"
        value={stats.completed}
        icon={<BiCheckCircle />}
        color="from-green-400 to-green-600"
      />
      <StatCard
        title="Pending"
        value={stats.pending}
        icon={<BiTime />}
        color="from-yellow-400 to-yellow-600"
      />
      <StatCard
        title="Failed"
        value={stats.failed}
        icon={<BiError />}
        color="from-red-400 to-red-600"
      />
    </div>
  );
}
