import { useEffect, useMemo, useState } from 'react';
import { FixedSizeList as List } from 'react-window';
import { fetchJobs, type Job } from '../services/api';
import { useUi } from '../contexts/UiContext';

const STATUS_OPTIONS = ['All', 'Pending', 'Running', 'Completed', 'Failed'];

const JobList = () => {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // UI state
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState(search);
  const [statusFilter, setStatusFilter] = useState('All');
  const [pageSize, setPageSize] = useState(10);
  const [page, setPage] = useState(1);
  const [sortBy, setSortBy] = useState<'created_at' | 'priority' | 'status' | 'type'>('created_at');
  const [sortDir, setSortDir] = useState<'desc' | 'asc'>('desc');
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [compact, setCompact] = useState(false);

  useEffect(() => {
    const loadJobs = async () => {
      try {
        const data = await fetchJobs();
        setJobs(data);
        setError(null);
      } catch (err) {
        setError('Failed to fetch jobs. Please try again later.');
      } finally {
        setIsLoading(false);
      }
    };

    loadJobs();
    const interval = setInterval(loadJobs, 5000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => setPage(1), [debouncedSearch, statusFilter, pageSize]);

  // debounce search input
  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search.trim()), 300);
    return () => clearTimeout(t);
  }, [search]);

  const filtered = useMemo(() => {
    const s = debouncedSearch.toLowerCase();
    return jobs.filter((j) => {
      if (statusFilter !== 'All' && j.status !== statusFilter) return false;
      if (!s) return true;
      return (
        j.id.toLowerCase().includes(s) ||
        j.type.toLowerCase().includes(s) ||
        (j.result && JSON.stringify(j.result).toLowerCase().includes(s))
      );
    });
  }, [jobs, debouncedSearch, statusFilter]);

  const sorted = useMemo(() => {
    const copy = [...filtered];
    copy.sort((a, b) => {
      let va: any = a[sortBy as keyof Job];
      let vb: any = b[sortBy as keyof Job];
      if (sortBy === 'created_at') {
        va = new Date(a.created_at).getTime();
        vb = new Date(b.created_at).getTime();
      }
      if (va === vb) return 0;
      if (sortDir === 'asc') return va > vb ? 1 : -1;
      return va < vb ? 1 : -1;
    });
    return copy;
  }, [filtered, sortBy, sortDir]);

  const totalPages = Math.max(1, Math.ceil(sorted.length / pageSize));
  const pageStart = (page - 1) * pageSize;
  const pageItems = sorted.slice(pageStart, pageStart + pageSize);

  const toggleSort = (col: typeof sortBy) => {
    if (sortBy === col) setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    else {
      setSortBy(col);
      setSortDir('desc');
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
    } catch (e) {
      // fallback
      const ta = document.createElement('textarea');
      ta.value = text;
      document.body.appendChild(ta);
      ta.select();
      document.execCommand('copy');
      document.body.removeChild(ta);
    }
  };

  if (isLoading) {
    return (
      <div className="p-4">
        <div className="animate-pulse flex space-x-4">
          <div className="flex-1 space-y-4 py-1">
            <div className="h-4 bg-gray-200 w-3/4" />
            <div className="space-y-2">
              <div className="h-4 bg-gray-200" />
              <div className="h-4 bg-gray-200 w-5/6" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <div className="bg-red-50 text-red-700 p-4">{error}</div>
      </div>
    );
  }

  const { screenshotMode } = useUi();

  return (
    <div className="p-4">
      <div className="flex items-center justify-between gap-4 mb-4">
        <h2 className="text-xl font-semibold text-gray-800">Jobs</h2>

        {!screenshotMode && (
          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600">Compact</label>
            <input type="checkbox" checked={compact} onChange={() => setCompact((c) => !c)} className="h-4 w-4" />
          </div>
        )}
      </div>

      {!screenshotMode && (
        <div className="flex flex-col md:flex-row md:items-center md:gap-4 mb-4">
          <div className="flex-1">
            <input
              placeholder="Search by ID, type or result..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full border border-gray-300 p-2 text-sm"
            />
          </div>

          <div className="flex items-center gap-2 mt-2 md:mt-0">
            <select className="border border-gray-300 p-2 text-sm" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
              {STATUS_OPTIONS.map((s) => (
                <option key={s} value={s}>{s}</option>
              ))}
            </select>
            <select className="border border-gray-300 p-2 text-sm" value={pageSize} onChange={(e) => setPageSize(parseInt(e.target.value))}>
              {[5,10,25,50].map(n => <option key={n} value={n}>{n} / page</option>)}
            </select>
            <button
              onClick={() => {
                const rows = [['id','type','status','priority','thread_demand','created_at','started_at','completed_at']].concat(
                  filtered.map(j => [j.id, j.type, j.status, String(j.priority), String(j.thread_demand), j.created_at || '', j.started_at || '', j.completed_at || ''])
                );
                const csv = rows.map(r => r.map(c => `"${String(c).replace(/"/g,'""')}"`).join(',')).join('\n');
                const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = `jobs_export_${new Date().toISOString()}.csv`;
                document.body.appendChild(a);
                a.click();
                a.remove();
                URL.revokeObjectURL(url);
              }}
              className="border border-gray-300 px-3 py-2 text-sm"
            >Export</button>
          </div>
        </div>
      )}

      <div className="overflow-x-auto">
        {sorted.length > 200 ? (
          // Virtualized list for large datasets
          <div className="bg-white shadow-sm border border-gray-200">
            <div className="grid grid-cols-9 bg-gray-100 text-gray-700 text-sm tracking-wider py-2 px-3">
              <div className="col-span-1">ID</div>
              <div className="col-span-1">Type</div>
              <div className="col-span-1">Status</div>
              <div className="col-span-1">Pri</div>
              <div className="col-span-1">Threads</div>
              <div className="col-span-2">Created</div>
              <div className="col-span-1">Started</div>
              <div className="col-span-1">Completed</div>
            </div>
            <List
              height={Math.min(600, pageSize * (compact ? 28 : 48))}
              itemCount={sorted.length}
              itemSize={compact ? 36 : 56}
              width="100%"
            >
              {({ index, style }) => {
                const job = sorted[index];
                return (
                  <div key={job.id} style={style} className="grid grid-cols-9 items-center border-b border-gray-100 px-3">
                    <div className="col-span-1 text-xs font-mono text-gray-700 truncate">{job.id}</div>
                    <div className="col-span-1 text-sm text-gray-900">{job.type}</div>
                    <div className="col-span-1 text-sm">{job.status}</div>
                    <div className="col-span-1 text-sm">{job.priority}</div>
                    <div className="col-span-1 text-sm">{job.thread_demand}</div>
                    <div className="col-span-2 text-sm text-gray-600">{new Date(job.created_at).toLocaleString()}</div>
                    <div className="col-span-1 text-sm text-gray-600">{job.started_at ? new Date(job.started_at).toLocaleString() : '-'}</div>
                    <div className="col-span-1 text-sm text-gray-600">{job.completed_at ? new Date(job.completed_at).toLocaleString() : '-'}</div>
                  </div>
                );
              }}
            </List>
          </div>
        ) : (
          <table className="min-w-full bg-white shadow-sm border border-gray-200">
            <thead>
              <tr className="bg-gray-100 text-gray-700 text-sm tracking-wider">
                <th className={`py-${compact? '1':'3'} px-4 text-left`}><button onClick={() => toggleSort('created_at')} className="flex items-center gap-2">ID</button></th>
                <th className={`py-${compact? '1':'3'} px-4 text-left`}><button onClick={() => toggleSort('type')} className="flex items-center gap-2">Type</button></th>
                <th className={`py-${compact? '1':'3'} px-4 text-left`}><button onClick={() => toggleSort('status')} className="flex items-center gap-2">Status</button></th>
                <th className={`py-${compact? '1':'3'} px-4 text-left`}><button onClick={() => toggleSort('priority')} className="flex items-center gap-2">Priority</button></th>
                <th className={`py-${compact? '1':'3'} px-4 text-left`}>Threads</th>
                <th className={`py-${compact? '1':'3'} px-4 text-left`}>Created</th>
                <th className={`py-${compact? '1':'3'} px-4 text-left`}>Started</th>
                <th className={`py-${compact? '1':'3'} px-4 text-left`}>Completed</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {pageItems.map((job) => (
                <>
                <tr key={job.id} className="hover:bg-gray-50">
                  <td className={`py-${compact? '1':'3'} px-4 font-mono text-sm text-gray-700`}> 
                    <div className="flex items-center gap-2">
                      <span className="truncate max-w-[18rem]">{job.id}</span>
                      <button onClick={() => copyToClipboard(job.id)} className="text-xs text-gray-500 hover:text-gray-700">Copy</button>
                      <button onClick={() => setExpandedId(expandedId === job.id ? null : job.id)} className="text-xs text-gray-500 hover:text-gray-700">Details</button>
                    </div>
                  </td>
                  <td className={`py-${compact? '1':'3'} px-4 text-gray-900`}>{job.type}</td>
                  <td className={`py-${compact? '1':'3'} px-4`}>
                    <span className={`px-2 text-sm font-medium ${job.status === 'Failed' ? 'text-red-700' : 'text-gray-800'}`}>
                      {job.status}
                    </span>
                  </td>
                  <td className={`py-${compact? '1':'3'} px-4 text-gray-900`}>{job.priority}</td>
                  <td className={`py-${compact? '1':'3'} px-4 text-gray-900`}>{job.thread_demand}</td>
                  <td className={`py-${compact? '1':'3'} px-4 text-gray-600`}>{new Date(job.created_at).toLocaleString()}</td>
                  <td className={`py-${compact? '1':'3'} px-4 text-gray-600`}>{job.started_at ? new Date(job.started_at).toLocaleString() : '-'}</td>
                  <td className={`py-${compact? '1':'3'} px-4 text-gray-600`}>{job.completed_at ? new Date(job.completed_at).toLocaleString() : '-'}</td>
                </tr>
                {expandedId === job.id && (
                  <tr key={`${job.id}-details`} className="bg-gray-50">
                    <td colSpan={8} className="p-4 text-sm text-gray-700">
                      <div className="mb-2 font-medium">Payload</div>
                      <pre className="whitespace-pre-wrap bg-white border border-gray-200 p-2 text-xs overflow-auto">{JSON.stringify(job.payload, null, 2)}</pre>
                      {job.result && (
                        <>
                          <div className="mt-3 mb-2 font-medium">Result</div>
                          <pre className="whitespace-pre-wrap bg-white border border-gray-200 p-2 text-xs overflow-auto">{JSON.stringify(job.result, null, 2)}</pre>
                        </>
                      )}
                    </td>
                  </tr>
                )}
                </>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {!screenshotMode && (
        <div className="flex items-center justify-between mt-4">
          <div className="text-sm text-gray-600">Showing {Math.min(sorted.length, pageStart+1)} - {Math.min(sorted.length, pageStart+pageSize)} of {sorted.length}</div>
          <div className="flex items-center gap-2">
            <button disabled={page<=1} onClick={() => setPage((p) => Math.max(1, p-1))} className="px-3 py-1 border border-gray-300 text-sm">Prev</button>
            <div className="px-3 py-1 text-sm">Page {page} / {totalPages}</div>
            <button disabled={page>=totalPages} onClick={() => setPage((p) => Math.min(totalPages, p+1))} className="px-3 py-1 border border-gray-300 text-sm">Next</button>
          </div>
        </div>
      )}
    </div>
  );
};

export default JobList;
