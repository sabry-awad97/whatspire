/**
 * API Key Revocation Dialog Component
 * Provides confirmation dialog for revoking API keys with optional reason
 */
import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Loader2 } from "lucide-react";
import { toast } from "sonner";

import { revokeAPIKeyMutation } from "@whatspire/hooks";
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
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import { Alert, AlertDescription } from "../ui/alert";

// ============================================================================
// Types
// ============================================================================

export interface RevokeAPIKeyDialogProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Callback when dialog open state changes */
  onOpenChange: (open: boolean) => void;
  /** API key to revoke */
  apiKey: {
    id: string;
    masked_key: string;
    description?: string | null;
    role: string;
  };
}

// ============================================================================
// Component
// ============================================================================

export function RevokeAPIKeyDialog({
  open,
  onOpenChange,
  apiKey,
}: RevokeAPIKeyDialogProps) {
  const [reason, setReason] = useState("");
  const apiClient = useApiClient();
  const queryClient = useQueryClient();

  const revokeMutation = useMutation(
    revokeAPIKeyMutation(apiClient, queryClient),
  );

  const handleRevoke = async () => {
    try {
      await revokeMutation.mutateAsync({
        id: apiKey.id,
        reason: reason.trim() || undefined,
      });

      toast.success("API Key revoked successfully");

      // Reset form and close dialog
      setReason("");
      onOpenChange(false);
    } catch (error) {
      toast.error(
        error instanceof Error
          ? error.message
          : "Failed to revoke API key. Please try again.",
      );
    }
  };

  const handleCancel = () => {
    setReason("");
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="glass-card-enhanced sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="text-2xl gradient-text">
            Revoke API Key
          </DialogTitle>
          <DialogDescription className="text-muted-foreground">
            This action cannot be undone. The API key will be immediately
            deactivated and all requests using it will fail.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Warning Alert */}
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              <strong>Warning:</strong> This will immediately revoke access for
              any applications using this key.
            </AlertDescription>
          </Alert>

          {/* Key Details */}
          <div className="glass-card p-4 rounded-lg space-y-2">
            <div className="text-sm">
              <span className="font-medium text-muted-foreground">Key:</span>{" "}
              <code className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
                {apiKey.masked_key}
              </code>
            </div>
            {apiKey.description && (
              <div className="text-sm">
                <span className="font-medium text-muted-foreground">
                  Description:
                </span>{" "}
                {apiKey.description}
              </div>
            )}
            <div className="text-sm">
              <span className="font-medium text-muted-foreground">Role:</span>{" "}
              <span className="capitalize">{apiKey.role}</span>
            </div>
          </div>

          {/* Reason Input */}
          <div className="space-y-2">
            <Label htmlFor="reason">
              Reason for Revocation{" "}
              <span className="text-muted-foreground">(optional)</span>
            </Label>
            <Textarea
              id="reason"
              placeholder="e.g., Key compromised, no longer needed, rotating keys..."
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={3}
              maxLength={500}
              disabled={revokeMutation.isPending}
              className="glass-card resize-none"
            />
            <p className="text-xs text-muted-foreground">
              This will be recorded in the audit log for future reference.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={handleCancel}
            disabled={revokeMutation.isPending}
            className="glass-card"
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleRevoke}
            disabled={revokeMutation.isPending}
            className="hover-glow-red"
          >
            {revokeMutation.isPending && (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            )}
            Revoke Key
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
