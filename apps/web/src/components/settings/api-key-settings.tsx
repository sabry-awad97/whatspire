/**
 * API Key Settings Component
 * Main component for API key management integrating list, create, and revoke functionality
 */
import { useState } from "react";
import { Key, Plus, RefreshCw } from "lucide-react";

import { Button } from "../ui/button";
import { APIKeyList } from "./api-key-list";
import { CreateAPIKeyDialog } from "./api-key-create-dialog";

// ============================================================================
// Component
// ============================================================================

export function APIKeySettings() {
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [filters, setFilters] = useState<{
    role?: "read" | "write" | "admin";
    status?: "active" | "revoked";
  }>({});
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  // Handle page change
  const handlePageChange = (newPage: number) => {
    setPage(newPage);
  };

  // Handle page size change
  const handlePageSizeChange = (newLimit: number) => {
    setLimit(newLimit);
    setPage(1); // Reset to first page when changing page size
  };

  // Handle filters change
  const handleFiltersChange = (newFilters: {
    role?: "read" | "write" | "admin";
    status?: "active" | "revoked";
  }) => {
    setFilters(newFilters);
    setPage(1); // Reset to first page when changing filters
  };

  // Handle refresh
  const handleRefresh = () => {
    setRefreshKey((prev) => prev + 1);
  };

  return (
    <div className="space-y-6">
      {/* Header Section */}
      <div className="glass-card-enhanced p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg glass-card bg-teal/10">
              <Key className="h-5 w-5 text-teal" />
            </div>
            <div>
              <h2 className="text-xl font-semibold">API Key Management</h2>
              <p className="text-sm text-muted-foreground">
                Create, manage, and revoke API keys for programmatic access
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleRefresh}
              className="glass-card hover-glow-teal"
              title="Refresh list"
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
            <Button
              onClick={() => setCreateDialogOpen(true)}
              className="glass-card hover-glow-teal"
            >
              <Plus className="mr-2 h-4 w-4" />
              Create API Key
            </Button>
          </div>
        </div>
      </div>

      {/* Info Banner */}
      <div className="glass-card p-4 border-2 border-teal/20 bg-teal/5 rounded-lg">
        <div className="flex items-start gap-3">
          <span className="text-2xl">ℹ️</span>
          <div className="flex-1">
            <h4 className="font-semibold text-teal mb-1">About API Keys</h4>
            <p className="text-sm text-muted-foreground">
              API keys provide programmatic access to the Whatspire API. Each
              key has a specific role (read, write, or admin) that determines
              its permissions. Keys are displayed only once during creation, so
              make sure to save them securely. You can revoke keys at any time
              if they are compromised.
            </p>
          </div>
        </div>
      </div>

      {/* API Key List */}
      <div className="glass-card-enhanced p-6">
        <APIKeyList
          key={refreshKey}
          filters={filters}
          page={page}
          limit={limit}
          onPageChange={handlePageChange}
          onPageSizeChange={handlePageSizeChange}
          onFiltersChange={handleFiltersChange}
        />
      </div>

      {/* Create API Key Dialog */}
      <CreateAPIKeyDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
      />
    </div>
  );
}
