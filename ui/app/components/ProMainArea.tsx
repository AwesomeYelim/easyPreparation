"use client";
import UpdateChecker from "./UpdateChecker";
import OnboardingChecklist from "./OnboardingChecklist";

export default function ProMainArea({ children }: { children: React.ReactNode }) {
  return (
    <main
      className="overflow-auto min-h-0 bg-pro-bg"
      style={{ gridColumn: "3", gridRow: "2" }}
    >
      <UpdateChecker />
      <div className="p-6 lg:p-8">
        <OnboardingChecklist />
        {children}
      </div>
    </main>
  );
}
