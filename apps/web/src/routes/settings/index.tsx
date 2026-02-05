import { createFileRoute } from "@tanstack/react-router";

import { Settings } from "@/components/settings/settings";

export const Route = createFileRoute("/settings/")({
  component: SettingsComponent,
});

function SettingsComponent() {
  return (
    <div className="min-h-screen network-bg">
      <div className="max-w-4xl mx-auto p-6">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold gradient-text">Settings</h1>
          <p className="text-muted-foreground">
            Manage your application preferences
          </p>
        </div>

        {/* Settings Content */}
        <Settings />
      </div>
    </div>
  );
}
