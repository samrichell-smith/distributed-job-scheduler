import { useState, useEffect } from 'react';
import { fetchJobs, type Job } from '../services/api';
import {
  BarChart, Bar,
  PieChart, Pie,
  LineChart, Line,
  XAxis, YAxis,
  Tooltip,
  CartesianGrid,
  ResponsiveContainer,
  Cell,
  Legend
} from 'recharts';

type JobTypeStats = {
  name: string;
  count: number;
};

type CompletionTimeStats = {
  type: string;
  avgTime: number;
};

type PriorityStats = {
  priority: number;
  count: number;
};

const COLORS = ['#6B7280', '#10B981', '#F59E0B', '#B91C1C'];

export default function Analytics() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadJobs = async () => {
      try {
        const fetchedJobs = await fetchJobs();
        setJobs(fetchedJobs);
      } catch (error) {
        console.error('Failed to fetch jobs:', error);
      } finally {
        setLoading(false);
      }
    };

    loadJobs();
    const interval = setInterval(loadJobs, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, []);

  const jobTypeData = jobs.reduce((acc: JobTypeStats[], job) => {
    const existingType = acc.find(item => item.name === job.type);
    if (existingType) {
      existingType.count++;
    } else {
      acc.push({ name: job.type, count: 1 });
    }
    return acc;
  }, []).sort((a, b) => b.count - a.count); // Sort by count descending

  const completionTimeData: CompletionTimeStats[] = Object.entries(
    jobs
      .filter(job => job.completed_at && job.started_at)
      .reduce((acc: { [key: string]: { total: number; count: number } }, job) => {
        const completionTime = new Date(job.completed_at!).getTime() - new Date(job.started_at!).getTime();
        if (!acc[job.type]) {
          acc[job.type] = { total: 0, count: 0 };
        }
        acc[job.type].total += completionTime;
        acc[job.type].count++;
        return acc;
      }, {})
  )
    .map(([type, data]) => ({
      type,
      avgTime: data.total / data.count
    }))
    .sort((a, b) => b.avgTime - a.avgTime); // Sort by average time descending

  const priorityData: PriorityStats[] = jobs.reduce((acc: PriorityStats[], job) => {
    const existingPriority = acc.find(item => item.priority === job.priority);
    if (existingPriority) {
      existingPriority.count++;
    } else {
      acc.push({ priority: job.priority, count: 1 });
    }
    return acc.sort((a, b) => a.priority - b.priority);
  }, []);

  const successRate = {
    success: jobs.filter(job => job.status === 'Completed').length,
    failed: jobs.filter(job => job.status === 'Failed').length,
    pending: jobs.filter(job => ['Pending', 'Running'].includes(job.status)).length,
  };

  const statusData = [
    { name: 'Completed', value: successRate.success },
    { name: 'Failed', value: successRate.failed },
    { name: 'In Progress', value: successRate.pending },
  ].filter(item => item.value > 0); // Only show non-zero values

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto p-8">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4"></div>
          <div className="grid grid-cols-2 gap-6">
            <div className="h-64 bg-gray-200 rounded"></div>
            <div className="h-64 bg-gray-200 rounded"></div>
            <div className="h-64 bg-gray-200 rounded"></div>
            <div className="h-64 bg-gray-200 rounded"></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-8">
      <h2 className="text-2xl font-bold mb-6 text-gray-800">Job Analytics</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Job Type Distribution */}
        <div className="bg-white p-6 shadow-sm">
          <h3 className="text-lg font-semibold mb-4 text-gray-700">Job Type Distribution</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={jobTypeData} margin={{ top: 10, right: 30, left: 0, bottom: 20 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis 
                dataKey="name" 
                angle={-45}
                textAnchor="end"
                height={60}
                interval={0}
              />
              <YAxis allowDecimals={false} />
              <Tooltip 
                cursor={{ fill: 'rgba(0, 0, 0, 0.1)' }}
                contentStyle={{ 
                  backgroundColor: 'rgba(255, 255, 255, 0.9)',
                  border: '1px solid #ccc',
                  borderRadius: '4px'
                }}
              />
              <Bar 
                dataKey="count" 
                fill="#374151"
                maxBarSize={50}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Success/Failure Rate */}
        <div className="bg-white p-6 shadow-sm">
          <h3 className="text-lg font-semibold mb-4 text-gray-700">Job Status Distribution</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={statusData}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }: { name: string, percent?: number }) => {
                  const percentage = percent ? (percent * 100) : 0;
                  return percentage > 3 ? `${name} ${percentage.toFixed(0)}%` : '';
                }}
                outerRadius={100}
                innerRadius={60}
                fill="#6B7280"
                dataKey="value"
              >
                {statusData.map((_entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>

        {/* Average Completion Time */}
        <div className="bg-white p-6 shadow-sm">
          <h3 className="text-lg font-semibold mb-4 text-gray-700">Average Completion Time by Type</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={completionTimeData} margin={{ top: 10, right: 30, left: 0, bottom: 20 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis 
                dataKey="type" 
                angle={-45}
                textAnchor="end"
                height={60}
                interval={0}
              />
              <YAxis 
                tickFormatter={(value) => `${(value / 1000).toFixed(1)}s`}
              />
              <Tooltip 
                cursor={{ fill: 'rgba(0, 0, 0, 0.1)' }}
                formatter={(value: number) => [`${(value / 1000).toFixed(2)}s`, 'Avg Time']}
                contentStyle={{ 
                  backgroundColor: 'rgba(255, 255, 255, 0.9)',
                  border: '1px solid #ccc',
                  borderRadius: '4px'
                }}
              />
              <Bar 
                dataKey="avgTime" 
                fill="#374151"
                maxBarSize={50}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Priority Distribution */}
        <div className="bg-white p-6 shadow-sm">
          <h3 className="text-lg font-semibold mb-4 text-gray-700">Job Priority Distribution</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={priorityData} margin={{ top: 10, right: 30, left: 0, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis 
                dataKey="priority" 
                allowDecimals={false}
              />
              <YAxis allowDecimals={false} />
              <Tooltip 
                cursor={{ stroke: '#666', strokeWidth: 1 }}
                contentStyle={{ 
                  backgroundColor: 'rgba(255, 255, 255, 0.9)',
                  border: '1px solid #ccc',
                  borderRadius: '4px'
                }}
                formatter={(value: number) => [`${value} jobs`, 'Count']}
              />
              <Line 
                type="monotone" 
                dataKey="count" 
                stroke="#374151"
                strokeWidth={2}
                dot={{ fill: '#374151', strokeWidth: 2 }}
                activeDot={{ r: 6, stroke: '#374151', strokeWidth: 2 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="mt-6 grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-white p-4 shadow-sm">
          <h4 className="text-sm font-medium text-gray-500">Total Jobs</h4>
          <p className="text-2xl font-bold text-gray-800">{jobs.length}</p>
        </div>
        <div className="bg-white p-4 shadow-sm">
          <h4 className="text-sm font-medium text-gray-500">Success Rate</h4>
          <p className="text-2xl font-bold text-gray-800">
            {jobs.length ? ((successRate.success / jobs.length) * 100).toFixed(1) : 0}%
          </p>
        </div>
        <div className="bg-white p-4 shadow-sm">
          <h4 className="text-sm font-medium text-gray-500">Avg Priority</h4>
          <p className="text-2xl font-bold text-gray-800">
            {jobs.length ? (jobs.reduce((sum, job) => sum + job.priority, 0) / jobs.length).toFixed(1) : 0}
          </p>
        </div>
        <div className="bg-white p-4 shadow-sm">
          <h4 className="text-sm font-medium text-gray-500">Active Jobs</h4>
          <p className="text-2xl font-bold text-gray-800">{successRate.pending}</p>
        </div>
      </div>
    </div>
  );
}
