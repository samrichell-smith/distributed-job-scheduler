const mockJobs = [
	{
		id: "1",
		type: "add_numbers",
		status: "Completed",
		priority: 1,
		thread_demand: 2,
		created_at: new Date(Date.now() - 1000 * 60 * 60).toISOString(),
		started_at: new Date(Date.now() - 1000 * 60 * 50).toISOString(),
		completed_at: new Date(Date.now() - 1000 * 60 * 45).toISOString(),
	},
	{
		id: "2",
		type: "large_array_sum",
		status: "Pending",
		priority: 2,
		thread_demand: 4,
		created_at: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
		started_at: undefined,
		completed_at: undefined,
	},
	{
		id: "3",
		type: "resize_image",
		status: "Failed",
		priority: 3,
		thread_demand: 1,
		created_at: new Date(Date.now() - 1000 * 60 * 10).toISOString(),
		started_at: new Date(Date.now() - 1000 * 60 * 9).toISOString(),
		completed_at: new Date(Date.now() - 1000 * 60 * 8).toISOString(),
	},
];

export default function JobList() {
	const jobs = mockJobs;

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
				<tbody>
					{jobs.map((job) => (
						<tr
							key={job.id}
							className="border-b hover:bg-blue-50 transition"
						>
							<td className="py-2 px-4 font-mono text-xs text-blue-700 truncate max-w-xs">
								{job.id}
							</td>
							<td className="py-2 px-4">{job.type}</td>
							<td className="py-2 px-4">
								<span
									className={`px-2 py-1 rounded text-xs font-semibold ${
										job.status === "Completed"
											? "bg-green-100 text-green-700"
											: job.status === "Pending"
											? "bg-yellow-100 text-yellow-700"
											: job.status === "Failed"
											? "bg-red-100 text-red-700"
											: "bg-gray-100 text-gray-700"
									}`}
								>
									{job.status}
								</span>
							</td>
							<td className="py-2 px-4">{job.priority}</td>
							<td className="py-2 px-4">{job.thread_demand}</td>
							<td className="py-2 px-4">
								{new Date(job.created_at).toLocaleString()}
							</td>
							<td className="py-2 px-4">
								{job.started_at
									? new Date(job.started_at).toLocaleString()
									: "-"}
							</td>
							<td className="py-2 px-4">
								{job.completed_at
									? new Date(job.completed_at).toLocaleString()
									: "-"}
							</td>
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}
