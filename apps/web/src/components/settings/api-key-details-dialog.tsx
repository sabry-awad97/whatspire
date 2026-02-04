/**
 * API Key Details Dialog Component
 * Displays comprehensive information about an API key including metadata and usage statistics
 */
import { useQuery } from "@tanstack/react-query";
import {
  Calendar,
  Key,
  Shield,
  Activity,
  Clock,
  AlertCircle,
} from "lucide-react";
import { format } from "date-fns";

import { getAPIKeyDetailsOptions } from "@whatspire/hooks";
import { useApiClient } from "@whatspire/hooks";

import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import { Badge } from "../ui/badge";
import { Skeleton } from "../ui/skeleton";
import { Alert, AlertDescription } from "../ui/alert";

// ============================================================================
// Types
// ============================================================================

export interface APIKeyDetailsDialogProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Callback when dialog open state changes */
  onOpenChange: (open: boolean) => void;
  /** API key ID to display details for */
  apiKeyId: string;
}

// ============================================================================
// Component
// ============================================================================

export function APIKeyDetailsDialog({
  open,
  onOpenChange,
  apiKeyId,
}: APIKeyDetailsDialogProps) {
  const apiClient = useApiClient();

  const { data, isLoading, error } = useQuery(
    getAPIKeyDetailsOptions(apiClient, apiKeyId),
  );

  const apiKey = data?.api_key;
  const usageStats = data?.usage_stats;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="glass-card-enhanced sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-2xl gradient-text flex items-center gap-2">
            <Key className="h-5 w-5" />
            API Key Details
          </DialogTitle>
          <DialogDescription className="text-muted-foreground">
            Comprehensive information about this API key including metadata and
            usage statistics
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Loading State */}
          {isLoading && (
            <div className="space-y-4">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-32 w-full" />
              <Skeleton className="h-24 w-full" />
            </div>
          )}

          {/* Error State */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Failed to load API key details. Please try again.
              </AlertDescription>
            </Alert>
          )}

          {/* Content */}
          {apiKey && (
            <>
              {/* Status Badge */}
              <div className="flex items-center justify-between">
                <Badge
                  variant={apiKey.is_active ? "default" : "destructive"}
                  className="text-sm px-3 py-1"
                >
                  {apiKey.is_active ? "Active" : "Revoked"}
                </Badge>
                <Badge
                  variant="outline"
                  className="text-sm px-3 py-1 capitalize"
                >
                  <Shield className="h-3 w-3 mr-1" />
                  {apiKey.role}
                </Badge>
              </div>

              {/* Key Information */}
              <div className="glass-card p-4 rounded-lg space-y-3">
                <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide">
                  Key Information
                </h3>

                <div className="space-y-2">
                  <div className="flex items-start justify-between">
                    <span className="text-sm font-medium text-muted-foreground">
                      ID:
                    </span>
                    <code className="text-xs font-mono bg-muted px-2 py-1 rounded">
                      {apiKey.id}
                    </code>
                  </div>

                  <div className="flex items-start justify-between">
                    <span className="text-sm font-medium text-muted-foreground">
                      Key:
                    </span>
                    <code className="text-xs font-mono bg-muted px-2 py-1 rounded">
                      {apiKey.masked_key}
                    </code>
                  </div>

                  {apiKey.description && (
                    <div className="flex items-start justify-between">
                      <span className="text-sm font-medium text-muted-foreground">
                        Description:
                      </span>
                      <span className="text-sm text-right max-w-[300px]">
                        {apiKey.description}
                      </span>
                    </div>
                  )}
                </div>
              </div>

              {/* Timestamps */}
              <div className="glass-card p-4 rounded-lg space-y-3">
                <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                  <Clock className="h-4 w-4" />
                  Timeline
                </h3>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                      <Calendar className="h-3 w-3" />
                      Created:
                    </span>
                    <span className="text-sm">
                      {format(new Date(apiKey.created_at), "PPpp")}
                    </span>
                  </div>

                  {apiKey.last_used_at && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                        <Activity className="h-3 w-3" />
                        Last Used:
                      </span>
                      <span className="text-sm">
                        {format(new Date(apiKey.last_used_at), "PPpp")}
                      </span>
                    </div>
                  )}

                  {!apiKey.last_used_at && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                        <Activity className="h-3 w-3" />
                        Last Used:
                      </span>
                      <span className="text-sm text-muted-foreground italic">
                        Never used
                      </span>
                    </div>
                  )}

                  {apiKey.revoked_at && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-destructive flex items-center gap-2">
                        <AlertCircle className="h-3 w-3" />
                        Revoked:
                      </span>
                      <span className="text-sm text-destructive">
                        {format(new Date(apiKey.revoked_at), "PPpp")}
                      </span>
                    </div>
                  )}
                </div>
              </div>

              {/* Revocation Details */}
              {apiKey.revoked_at && (
                <div className="glass-card p-4 rounded-lg space-y-3 border-destructive/50">
                  <h3 className="font-semibold text-sm text-destructive uppercase tracking-wide">
                    Revocation Details
                  </h3>

                  <div className="space-y-2">
                    {apiKey.revoked_by && (
                      <div className="flex items-start justify-between">
                        <span className="text-sm font-medium text-muted-foreground">
                          Revoked By:
                        </span>
                        <span className="text-sm">{apiKey.revoked_by}</span>
                      </div>
                    )}

                    {apiKey.revocation_reason && (
                      <div className="flex items-start justify-between">
                        <span className="text-sm font-medium text-muted-foreground">
                          Reason:
                        </span>
                        <span className="text-sm text-right max-w-[300px]">
                          {apiKey.revocation_reason}
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Usage Statistics */}
              {usageStats && (
                <div className="glass-card p-4 rounded-lg space-y-3">
                  <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                    <Activity className="h-4 w-4" />
                    Usage Statistics
                  </h3>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="glass-card p-3 rounded-lg text-center">
                      <div className="text-2xl font-bold gradient-text">
                        {usageStats.total_requests.toLocaleString()}
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        Total Requests
                      </div>
                    </div>

                    <div className="glass-card p-3 rounded-lg text-center">
                      <div className="text-2xl font-bold gradient-text">
                        {usageStats.last_7_days.toLocaleString()}
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        Last 7 Days
                      </div>
                    </div>
                  </div>

                  {usageStats.total_requests === 0 && (
                    <p className="text-xs text-muted-foreground text-center italic">
                      No usage recorded yet
                    </p>
                  )}
                </div>
              )}
            </>
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            className="glass-card"
          >
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
