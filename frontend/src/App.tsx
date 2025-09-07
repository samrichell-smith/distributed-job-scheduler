
import DashboardLayout from "./layouts/DashboardLayout";
import JobList from "./components/JobList";
import Stats from "./components/Stats";
import JobSubmitForm from "./components/JobSubmitForm";
import Analytics from "./pages/Analytics";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { useState } from "react";
import "./index.css";


function DashboardHome() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  return (
    <div className="max-w-7xl mx-auto w-full">
      <div className="flex justify-between items-center">
        <Stats />
        <div className="ml-4">
          <JobSubmitForm onSubmit={() => setRefreshTrigger(prev => prev + 1)} />
        </div>
      </div>
      <div className="mt-8">
        <JobList key={refreshTrigger} />
      </div>
    </div>
  );
}

function App() {
  return (
    <Router>
      <DashboardLayout>
        <Routes>
          <Route path="/" element={<DashboardHome />} />
          <Route path="/analytics" element={<Analytics />} />
        </Routes>
      </DashboardLayout>
    </Router>
  );
}

export default App;
