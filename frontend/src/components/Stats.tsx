import { useState, useEffect } from 'react';
import { type JobStats, fetchJobs } from '../services/api';
import StatCard from './StatCard';
import { BiTask, BiCheckCircle, BiError, BiTime, BiChip } from 'react-icons/bi';
import { useUi } from '../contexts/UiContext';

export default function Stats() {
  const [stats, setStats] = useState<JobStats & { 
    total: number;
    averageCompletion?: number;
    totalThreads: number;
  }>({
    completed: 0,
    pending: 0,
    failed: 0,
    total: 0,
    totalThreads: 0,
  });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadStats = async () => {
      try {
        const jobs = await fetchJobs();
        console.log('All jobs:', jobs.map(j => `${j.id}: ${j.status}`));
        
        // Filter jobs by status, matching backend state exactly
        const completed = jobs.filter(j => j.status === 'Completed');
        const pending = jobs.filter(j => j.status === 'Pending');
        const running = jobs.filter(j => j.status === 'Running');
        const failed = jobs.filter(j => j.status === 'Failed');
        
        console.log('Stats:', {
          completed: completed.length,
          pending: pending.length,
          running: running.length,
          failed: failed.length
        });
        
        // Calculate average completion time for recently completed jobs (last 24h)
        const oneDayAgo = new Date(Date.now() - 24 * 60 * 60 * 1000);
        const recentCompleted = completed
          .filter(j => j.completed_at && new Date(j.completed_at) > oneDayAgo)
          .filter(j => j.completed_at && j.started_at)
          .map(j => new Date(j.completed_at!).getTime() - new Date(j.started_at!).getTime());
        
        const averageCompletion = recentCompleted.length > 0
          ? recentCompleted.reduce((a, b) => a + b, 0) / recentCompleted.length
          : undefined;

        // Calculate active thread demand (only from Running jobs)
        const activeThreads = running.reduce((sum, job) => sum + job.thread_demand, 0);

        setStats({
          completed: completed.length,
          pending: pending.length + running.length,  // Active jobs include both pending and running
          failed: failed.length,
          total: jobs.length,
          averageCompletion,
          totalThreads: activeThreads,  // Only count threads from running jobs
        });
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

  const { screenshotMode } = useUi();

  if (isLoading) {
    return (
      <div className={screenshotMode ? 'grid grid-cols-4 gap-8 p-8' : 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 p-4'}>
        {[...Array(4)].map((_, i) => (
          <div key={i} className="animate-pulse">
            <div className="bg-gray-200 h-[110px]" />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className={screenshotMode ? 'grid grid-cols-4 gap-8 p-8' : 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 p-4'}>
      <StatCard title="Total Jobs" value={stats.total} icon={<BiTask />} large={screenshotMode} />
      <StatCard title="Running" value={stats.pending} icon={<BiTime />} large={screenshotMode} />
      <StatCard title="Completed" value={`${stats.completed} (${((stats.completed / Math.max(1, stats.total)) * 100).toFixed(1)}%)`} icon={<BiCheckCircle />} large={screenshotMode} />
      <StatCard title="Failed" value={stats.failed} icon={<BiError />} large={screenshotMode} />
      <StatCard title="Total Threads" value={stats.totalThreads} icon={<BiChip />} large={screenshotMode} />
      {stats.averageCompletion && (
        <StatCard title="Avg. Completion Time" value={`${(stats.averageCompletion / 1000).toFixed(2)}s`} icon={<BiTime />} large={screenshotMode} />
      )}
    </div>
  );
}
