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

export async function fetchJobs(): Promise<Job[]> {
  console.log('Fetching jobs from:', `${API_URL}/jobs`);
  try {
    const response = await fetch(`${API_URL}/jobs`, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      },
      mode: 'cors',
      cache: 'no-cache',
      credentials: 'same-origin'
    });
    console.log('Response status:', response.status);
    console.log('Response headers:', Object.fromEntries(response.headers.entries()));
    
    if (!response.ok) {
      throw new Error(`Failed to fetch jobs: ${response.status} ${response.statusText}`);
    }
    
    const data = await response.json();
    console.log('Fetched jobs:', data);
    return data;
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
