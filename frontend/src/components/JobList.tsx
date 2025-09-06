import { useState, useEffect } from 'react';
import { fetchJobs, type Job } from '../services/api';

const JobList = () => {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadJobs = async () => {
      console.log('JobList: Starting to load jobs...');
      try {
        const data = await fetchJobs();
        console.log('JobList: Successfully loaded jobs:', data);
        setJobs(data);
        setError(null);
      } catch (err) {
        console.error('JobList: Error loading jobs:', err);
        setError('Failed to fetch jobs. Please try again later.');
      } finally {
        console.log('JobList: Finished loading attempt, setting isLoading to false');
        setIsLoading(false);
      }
    };

    loadJobs();
    // Poll for updates every 5 seconds
    const interval = setInterval(loadJobs, 5000);
    return () => clearInterval(interval);
  }, []);

  if (isLoading) {
    return (
      <div className="p-4">
        <div className="animate-pulse flex space-x-4">
          <div className="flex-1 space-y-4 py-1">
            <div className="h-4 bg-gray-200 rounded w-3/4" />
            <div className="space-y-2">
              <div className="h-4 bg-gray-200 rounded" />
              <div className="h-4 bg-gray-200 rounded w-5/6" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="bg-red-50 text-red-700 p-4 rounded-lg">{error}</div>
      </div>
    );
  }

  return (
    <div className="overflow-x-auto p-4">
      <h2 className="text-xl font-semibold mb-4 text-gray-800">All Jobs</h2>
      <table className="min-w-full bg-white rounded-xl shadow-lg border border-gray-200">
        <thead>
          <tr className="bg-blue-50 text-blue-900 text-sm uppercase tracking-wider">
            <th className="py-3 px-4 text-left">ID</th>
            <th className="py-3 px-4 text-left">Type</th>
            <th className="py-3 px-4 text-left">Status</th>
            <th className="py-3 px-4 text-left">Priority</th>
            <th className="py-3 px-4 text-left">Threads</th>
            <th className="py-3 px-4 text-left">Created</th>
            <th className="py-3 px-4 text-left">Started</th>
            <th className="py-3 px-4 text-left">Completed</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200">
          {jobs.map((job) => (
            <tr key={job.id} className="hover:bg-gray-50">
              <td className="py-3 px-4 font-mono text-sm text-blue-600">{job.id}</td>
              <td className="py-3 px-4 text-gray-900">{job.type}</td>
              <td className="py-3 px-4">
                <span
                  className={`px-2 py-1 rounded-full text-sm font-medium ${
                    job.status === 'Completed'
                      ? 'bg-green-100 text-green-800'
                      : job.status === 'Running'
                      ? 'bg-blue-100 text-blue-800'
                      : job.status === 'Failed'
                      ? 'bg-red-100 text-red-800'
                      : 'bg-yellow-100 text-yellow-800'
                  }`}
                >
                  {job.status}
                </span>
              </td>
              <td className="py-3 px-4 text-gray-900">{job.priority}</td>
              <td className="py-3 px-4 text-gray-900">{job.thread_demand}</td>
              <td className="py-3 px-4 text-gray-600">
                {new Date(job.created_at).toLocaleString()}
              </td>
              <td className="py-3 px-4 text-gray-600">
                {job.started_at ? new Date(job.started_at).toLocaleString() : '-'}
              </td>
              <td className="py-3 px-4 text-gray-600">
                {job.completed_at ? new Date(job.completed_at).toLocaleString() : '-'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default JobList;
