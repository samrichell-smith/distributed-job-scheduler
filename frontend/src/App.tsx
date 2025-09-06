import DashboardLayout from "./layouts/DashboardLayout";
import JobList from "./components/JobList";
import StatCard from "./components/StatCard";
import { FaCheckCircle, FaHourglassHalf, FaExclamationTriangle } from "react-icons/fa";
import "./index.css";

function App() {
  // Example stats for demo
  const stats = [
    {
      title: "Completed Jobs",
      value: 128,
      icon: <FaCheckCircle className="text-green-500" />,
      color: "bg-green-50 border-green-200",
    },
    {
      title: "Pending Jobs",
      value: 7,
      icon: <FaHourglassHalf className="text-yellow-500" />,
      color: "bg-yellow-50 border-yellow-200",
    },
    {
      title: "Failed Jobs",
      value: 3,
      icon: <FaExclamationTriangle className="text-red-500" />,
      color: "bg-red-50 border-red-200",
    },
  ];

  return (
    <DashboardLayout>
      <div className="max-w-7xl mx-auto w-full">
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6 mb-8">
          {stats.map((stat) => (
            <StatCard key={stat.title} {...stat} />
          ))}
        </div>
        <JobList />
      </div>
    </DashboardLayout>
  );
}

export default App;
