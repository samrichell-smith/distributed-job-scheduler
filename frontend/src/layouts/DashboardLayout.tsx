import TopTabs from "../components/TopTabs";
import Header from "../components/Header";
import type { ReactNode } from "react";

export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <TopTabs />
      <main className="max-w-7xl mx-auto p-6">{children}</main>
    </div>
  );
}
