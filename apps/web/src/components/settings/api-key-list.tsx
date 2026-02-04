/**
 * API Key List Component
 * Displays paginated list of API keys with filtering and actions
 */
import { useState } from "react";
import { Copy, Eye, EyeOff, Key, Loader2, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { useApiClient, useAPIKeys } from "@whatspire/hooks";
import type { APIKey } from "@whatspire/schema";

import { cn } from "@/lib/utils";
import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
  PaginationEllipsis,
} from "../ui/pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";
import { RevokeAPIKeyDialog } from "./api-key-revoke-dialog";

// ============================================================================
// Types
// ============================================================================

export interface APIKeyListProps {
  /** Optional filters for role and status */
  filters?: {
    role?: "read" | "write" | "admin";
    status?: "active" | "revoked";
  };
  /** Current page number */
  page?: number;
  /** Items per page */
  limit?: number;
  /** Callback when page changes */
  onPageChange?: (page: number) => void;
  /** Callback when page size changes */
  onPageSizeChange?: (limit: number) => void;
  /** Callback when filters change */
  onFiltersChange?: (filters: {
    role?: "read" | "write" | "admin";
    status?: "active" | "revoked";
  }) => void;
}

// ============================================================================
// Component
// ============================================================================

export function APIKeyList({
  filters,
  page = 1,
  limit = 50,
  onPageChange,
  onPageSizeChange,
  onFiltersChange,
}: APIKeyListProps) {
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set());
  const [revokeDialogOpen, setRevokeDialogOpen] = useState(false);
  const [selectedKey, setSelectedKey] = useState<APIKey | null>(null);

  const apiClient = useApiClient();

  // Fetch API keys with filters and pagination
  const { data, isLoading, error } = useAPIKeys(apiClient, {
    page,
    limit,
    ...filters,
  });

  // Toggle key visibility
  const toggleKeyVisibility = (keyId: string) => {
    setVisibleKeys((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(keyId)) {
        newSet.delete(keyId);
      } else {
        newSet.add(keyId);
        // Auto-hide after 10 seconds
        setTimeout(() => {
          setVisibleKeys((current) => {
            const updated = new Set(current);
            updated.delete(keyId);
            return updated;
          });
        }, 10000);
      }
      return newSet;
    });
  };

  // Copy key to clipboard
  const handleCopyKey = async (maskedKey: string) => {
    try {
      await navigator.clipboard.writeText(maskedKey);
      toast.success("API Key copied to clipboard");
    } catch (error) {
      toast.error("Failed to copy to clipboard");
      console.error("Copy failed:", error);
    }
  };

  // Open revoke dialog
  const handleRevokeClick = (apiKey: APIKey) => {
    setSelectedKey(apiKey);
    setRevokeDialogOpen(true);
  };

  // Format date for display
  const formatDate = (dateString: string | null | undefined) => {
    if (!dateString) return "Never";
    return new Date(dateString).toLocaleString();
  };

  // Get role badge variant
  const getRoleBadgeVariant = (role: string) => {
    switch (role) {
      case "admin":
        return "destructive";
      case "write":
        return "default";
      case "read":
        return "secondary";
      default:
        return "outline";
    }
  };

  // Get role icon
  const getRoleIcon = (role: string) => {
    switch (role) {
      case "admin":
        return "🔑";
      case "write":
        return "✏️";
      case "read":
        return "👁️";
      default:
        return "🔒";
    }
  };

  // Generate page numbers for pagination
  const generatePageNumbers = () => {
    if (!data?.pagination) return [];

    const { page: currentPage, total_pages } = data.pagination;
    const pages: (number | "ellipsis")[] = [];

    if (total_pages <= 7) {
      // Show all pages if 7 or fewer
      for (let i = 1; i <= total_pages; i++) {
        pages.push(i);
      }
    } else {
      // Always show first page
      pages.push(1);

      if (currentPage > 3) {
        pages.push("ellipsis");
      }

      // Show pages around current page
      const start = Math.max(2, currentPage - 1);
      const end = Math.min(total_pages - 1, currentPage + 1);

      for (let i = start; i <= end; i++) {
        pages.push(i);
      }

      if (currentPage < total_pages - 2) {
        pages.push("ellipsis");
      }

      // Always show last page
      pages.push(total_pages);
    }

    return pages;
  };

  // Handle page change
  const handlePageChange = (newPage: number) => {
    if (onPageChange) {
      onPageChange(newPage);
    }
  };

  // Handle page size change
  const handlePageSizeChange = (newLimit: string) => {
    if (onPageSizeChange) {
      onPageSizeChange(Number(newLimit));
    }
  };

  // Handle role filter change
  const handleRoleFilterChange = (value: string) => {
    if (onFiltersChange) {
      onFiltersChange({
        ...filters,
        role:
          value === "all" ? undefined : (value as "read" | "write" | "admin"),
      });
    }
  };

  // Handle status filter change
  const handleStatusFilterChange = (value: string) => {
    if (onFiltersChange) {
      onFiltersChange({
        ...filters,
        status: value === "all" ? undefined : (value as "active" | "revoked"),
      });
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center py-12 space-y-4">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="text-sm text-muted-foreground">Loading API keys...</p>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12 space-y-4">
        <div className="glass-card p-6 text-center max-w-md">
          <p className="text-destructive font-medium mb-2">
            Failed to load API keys
          </p>
          <p className="text-sm text-muted-foreground mb-4">
            {error instanceof Error ? error.message : "Unknown error occurred"}
          </p>
        </div>
      </div>
    );
  }

  // Empty state
  if (!data || data.api_keys.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 space-y-2">
        <Key className="h-12 w-12 text-muted-foreground/50" />
        <p className="text-sm text-muted-foreground">No API keys found</p>
        <p className="text-xs text-muted-foreground">
          {filters?.role || filters?.status
            ? "Try adjusting your filters"
            : "Create your first API key to get started"}
        </p>
      </div>
    );
  }

  return (
    <>
      {/* Filter Controls */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 mb-6">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-muted-foreground">
            Filters:
          </span>
        </div>

        {/* Role Filter */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Role:</span>
          <Select
            value={filters?.role || "all"}
            onValueChange={handleRoleFilterChange}
          >
            <SelectTrigger className="glass-card w-[140px]">
              <SelectValue placeholder="All roles" />
            </SelectTrigger>
            <SelectContent className="glass-card-enhanced">
              <SelectItem value="all">All roles</SelectItem>
              <SelectItem value="read">
                <span className="flex items-center gap-2">
                  <span>👁️</span>
                  <span>Read</span>
                </span>
              </SelectItem>
              <SelectItem value="write">
                <span className="flex items-center gap-2">
                  <span>✏️</span>
                  <span>Write</span>
                </span>
              </SelectItem>
              <SelectItem value="admin">
                <span className="flex items-center gap-2">
                  <span>🔑</span>
                  <span>Admin</span>
                </span>
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Status Filter */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Status:</span>
          <Select
            value={filters?.status || "all"}
            onValueChange={handleStatusFilterChange}
          >
            <SelectTrigger className="glass-card w-[140px]">
              <SelectValue placeholder="All statuses" />
            </SelectTrigger>
            <SelectContent className="glass-card-enhanced">
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value="active">
                <span className="flex items-center gap-2">
                  <span className="text-emerald">✓</span>
                  <span>Active</span>
                </span>
              </SelectItem>
              <SelectItem value="revoked">
                <span className="flex items-center gap-2">
                  <span className="text-destructive">✕</span>
                  <span>Revoked</span>
                </span>
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Clear Filters Button */}
        {(filters?.role || filters?.status) && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onFiltersChange?.({})}
            className="glass-card text-muted-foreground hover:text-foreground"
          >
            Clear filters
          </Button>
        )}
      </div>

      <div className="glass-card rounded-xl overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[200px]">Key</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Last Used</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data.api_keys.map((apiKey) => {
              const isVisible = visibleKeys.has(apiKey.id);
              const isActive = apiKey.is_active;

              return (
                <TableRow
                  key={apiKey.id}
                  className={cn(
                    "transition-all",
                    !isActive && "opacity-60 bg-muted/20",
                  )}
                >
                  {/* Masked Key */}
                  <TableCell className="font-mono text-xs">
                    <div className="flex items-center gap-2">
                      <code
                        className={cn(
                          "px-2 py-1 rounded bg-muted/50 transition-all",
                          isVisible && "bg-primary/10",
                        )}
                      >
                        {isVisible
                          ? apiKey.masked_key
                          : apiKey.masked_key.replace(/./g, "•")}
                      </code>
                      <Button
                        variant="ghost"
                        size="icon-xs"
                        onClick={() => toggleKeyVisibility(apiKey.id)}
                        className="glass-card"
                        title={isVisible ? "Hide key" : "Show key"}
                      >
                        {isVisible ? (
                          <EyeOff className="h-3 w-3" />
                        ) : (
                          <Eye className="h-3 w-3" />
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-xs"
                        onClick={() => handleCopyKey(apiKey.masked_key)}
                        className="glass-card"
                        title="Copy to clipboard"
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                  </TableCell>

                  {/* Role */}
                  <TableCell>
                    <Badge
                      variant={getRoleBadgeVariant(apiKey.role)}
                      className="capitalize"
                    >
                      <span className="mr-1">{getRoleIcon(apiKey.role)}</span>
                      {apiKey.role}
                    </Badge>
                  </TableCell>

                  {/* Description */}
                  <TableCell className="max-w-[300px]">
                    <p
                      className="text-sm truncate"
                      title={apiKey.description || ""}
                    >
                      {apiKey.description || (
                        <span className="text-muted-foreground italic">
                          No description
                        </span>
                      )}
                    </p>
                  </TableCell>

                  {/* Status */}
                  <TableCell>
                    {isActive ? (
                      <Badge variant="outline" className="text-emerald">
                        <span className="mr-1">✓</span>
                        Active
                      </Badge>
                    ) : (
                      <Badge variant="destructive">
                        <span className="mr-1">✕</span>
                        Revoked
                      </Badge>
                    )}
                  </TableCell>

                  {/* Created Date */}
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(apiKey.created_at)}
                  </TableCell>

                  {/* Last Used Date */}
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDate(apiKey.last_used_at)}
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="text-right">
                    {isActive ? (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRevokeClick(apiKey)}
                        className="text-destructive hover:text-destructive hover-glow-destructive"
                      >
                        <Trash2 className="mr-2 h-3 w-3" />
                        Revoke
                      </Button>
                    ) : (
                      <span className="text-xs text-muted-foreground">
                        Revoked {formatDate(apiKey.revoked_at)}
                      </span>
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>

      {/* Pagination Info */}
      {data.pagination && (
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4 mt-6">
          {/* Page Size Selector */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">Show</span>
            <Select
              value={limit.toString()}
              onValueChange={handlePageSizeChange}
            >
              <SelectTrigger className="glass-card w-[80px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="glass-card-enhanced">
                <SelectItem value="10">10</SelectItem>
                <SelectItem value="25">25</SelectItem>
                <SelectItem value="50">50</SelectItem>
                <SelectItem value="100">100</SelectItem>
              </SelectContent>
            </Select>
            <span className="text-sm text-muted-foreground">
              of {data.pagination.total} API keys
            </span>
          </div>

          {/* Pagination Controls */}
          {data.pagination.total_pages > 1 && (
            <Pagination>
              <PaginationContent className="glass-card rounded-lg p-1">
                {/* Previous Button */}
                <PaginationItem>
                  <PaginationPrevious
                    href="#"
                    onClick={(e) => {
                      e.preventDefault();
                      if (data.pagination.page > 1) {
                        handlePageChange(data.pagination.page - 1);
                      }
                    }}
                    className={cn(
                      data.pagination.page === 1 &&
                        "pointer-events-none opacity-50",
                    )}
                  />
                </PaginationItem>

                {/* Page Numbers */}
                {generatePageNumbers().map((pageNum, index) => (
                  <PaginationItem key={index}>
                    {pageNum === "ellipsis" ? (
                      <PaginationEllipsis />
                    ) : (
                      <PaginationLink
                        href="#"
                        onClick={(e) => {
                          e.preventDefault();
                          handlePageChange(pageNum);
                        }}
                        isActive={pageNum === data.pagination.page}
                      >
                        {pageNum}
                      </PaginationLink>
                    )}
                  </PaginationItem>
                ))}

                {/* Next Button */}
                <PaginationItem>
                  <PaginationNext
                    href="#"
                    onClick={(e) => {
                      e.preventDefault();
                      if (data.pagination.page < data.pagination.total_pages) {
                        handlePageChange(data.pagination.page + 1);
                      }
                    }}
                    className={cn(
                      data.pagination.page === data.pagination.total_pages &&
                        "pointer-events-none opacity-50",
                    )}
                  />
                </PaginationItem>
              </PaginationContent>
            </Pagination>
          )}
        </div>
      )}

      {/* Revoke Dialog */}
      {selectedKey && (
        <RevokeAPIKeyDialog
          open={revokeDialogOpen}
          onOpenChange={setRevokeDialogOpen}
          apiKey={selectedKey}
        />
      )}
    </>
  );
}
