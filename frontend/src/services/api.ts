type JobType = 'AddNumbers' | 'ReverseString' | 'ResizeImage' | 'LargeArraySum';
type JobStatus = 'Pending' | 'Running' | 'Completed' | 'Failed';

interface Job {
  id: string;
  name: string;
  type: JobType;
  status: JobStatus;
  priority: number;
  thread_demand: number;
  payload: any;
  result?: any;
  created_at: string;
  started_at?: string;
  completed_at?: string;
}

interface JobStats {
  completed: number;
  pending: number;
  failed: number;
}

const API_URL = "http://localhost:8080";

async function fetchCurrentJobs(): Promise<Job[]> {
  const response = await fetch(`${API_URL}/jobs`, {
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
  });
  if (!response.ok) {
    throw new Error(`Failed to fetch current jobs: ${response.status} ${response.statusText}`);
  }
  return response.json();
}

async function fetchHistoricalJobs(): Promise<Job[]> {
  const response = await fetch(`${API_URL}/db/jobs`, {
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
  });
  if (!response.ok) {
    throw new Error(`Failed to fetch historical jobs: ${response.status} ${response.statusText}`);
  }
  return response.json();
}

export async function fetchJobs(): Promise<Job[]> {
  console.log('Fetching all jobs...');
  try {
    // Fetch both current and historical jobs in parallel
    const [currentJobs, historicalJobs] = await Promise.all([
      fetchCurrentJobs(),
      fetchHistoricalJobs()
    ]);

    console.log('Current jobs:', currentJobs);
    console.log('Historical jobs:', historicalJobs);

    // Combine and deduplicate jobs based on ID
    const jobMap = new Map<string, Job>();
    
    // Current jobs take precedence
    currentJobs.forEach(job => jobMap.set(job.id, job));
    historicalJobs.forEach(job => {
      if (!jobMap.has(job.id)) {
        jobMap.set(job.id, job);
      }
    });

    const allJobs = Array.from(jobMap.values());
    // Sort by creation date, newest first
    allJobs.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

    return allJobs;
  } catch (error) {
    console.error('Error fetching jobs:', error);
    throw error;
  }
}

export async function fetchJob(id: string): Promise<Job> {
  const response = await fetch(`${API_URL}/jobs/${id}`);
  if (!response.ok) {
    throw new Error('Failed to fetch job');
  }
  return response.json();
}

export async function submitJob(job: {
  type: string;
  priority: number;
  thread_demand: number;
  payload: any;
}): Promise<Job> {
  const response = await fetch(`${API_URL}/jobs`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(job),
  });
  if (!response.ok) {
    throw new Error('Failed to submit job');
  }
  return response.json();
}

export async function fetchJobStats(): Promise<JobStats> {
  const jobs = await fetchJobs();
  return {
    completed: jobs.filter(job => job.status === 'Completed').length,
    pending: jobs.filter(job => job.status === 'Pending' || job.status === 'Running').length,
    failed: jobs.filter(job => job.status === 'Failed').length,
  };
}

export type { Job, JobStats };
