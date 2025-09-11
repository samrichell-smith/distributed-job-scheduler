-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    priority INT NOT NULL,
    thread_demand INT NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result JSONB,
    worker_id VARCHAR(255)
);

-- Create metrics table
CREATE TABLE IF NOT EXISTS job_metrics (
    id SERIAL PRIMARY KEY,
    job_id VARCHAR(255) REFERENCES jobs(id),
    metric_name VARCHAR(50) NOT NULL,
    metric_value FLOAT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_job_metrics_job_id ON job_metrics(job_id);
CREATE INDEX IF NOT EXISTS idx_job_metrics_name ON job_metrics(metric_name);
