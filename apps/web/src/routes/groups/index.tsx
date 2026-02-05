import { createFileRoute } from "@tanstack/react-router";
import { toast } from "sonner";

import { GroupList } from "@/components/groups/group-list";

export const Route = createFileRoute("/groups/")({
  component: GroupsComponent,
});

function GroupsComponent() {
  const handleSync = () => {
    toast.success("Groups synced successfully");
  };

  return (
    <div className="h-screen network-bg flex flex-col">
      <div className="flex-1 max-w-7xl mx-auto w-full p-6 flex flex-col">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold gradient-text">Groups</h1>
          <p className="text-muted-foreground">Manage your WhatsApp groups</p>
        </div>

        {/* Group List */}
        <div className="flex-1 glass-card-enhanced rounded-lg overflow-hidden">
          <GroupList onSync={handleSync} />
        </div>
      </div>
    </div>
  );
}
