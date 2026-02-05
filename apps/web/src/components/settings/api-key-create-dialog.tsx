import { useForm } from "@tanstack/react-form";
import { Check, Copy, Key, Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { useApiClient, useCreateAPIKey } from "@whatspire/hooks";
import type { CreateAPIKeyResponse } from "@whatspire/schema";

import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "../ui/field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Textarea } from "../ui/textarea";

// ============================================================================
// Types
// ============================================================================

type DialogStep = "form" | "display-key";

export interface CreateAPIKeyDialogProps {
  /** Whether the dialog is open */
  open?: boolean;
  /** Callback when dialog open state changes */
  onOpenChange?: (open: boolean) => void;
}

// ============================================================================
// Component
// ============================================================================

export function CreateAPIKeyDialog({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
}: CreateAPIKeyDialogProps = {}) {
  const [internalOpen, setInternalOpen] = useState(false);
  const [step, setStep] = useState<DialogStep>("form");
  const [createdKey, setCreatedKey] = useState<CreateAPIKeyResponse | null>(
    null,
  );
  const [copied, setCopied] = useState(false);

  // Use controlled or uncontrolled state
  const open = controlledOpen !== undefined ? controlledOpen : internalOpen;
  const setOpen =
    controlledOnOpenChange !== undefined
      ? controlledOnOpenChange
      : setInternalOpen;

  const apiClient = useApiClient();

  const createAPIKey = useCreateAPIKey(apiClient, {
    onSuccess: (response) => {
      setCreatedKey(response);
      setStep("display-key");

      // Save the API key to localStorage for use in sessions
      localStorage.setItem("whatspire_api_token", response.plain_key);

      toast.success("API Key created successfully");
    },
    onError: (error) => {
      toast.error(
        error instanceof Error ? error.message : "Failed to create API Key",
      );
    },
  });

  const form = useForm({
    defaultValues: {
      role: "read" as "read" | "write" | "admin",
      description: "",
    },
    onSubmit: async ({ value }) => {
      createAPIKey.mutate({
        role: value.role,
        description: value.description || undefined,
      });
    },
  });

  const handleReset = () => {
    setStep("form");
    setCreatedKey(null);
    setCopied(false);
    form.reset();
  };

  const handleClose = () => {
    setOpen(false);
    setTimeout(handleReset, 300);
  };

  const handleCopyKey = async () => {
    if (!createdKey?.plain_key) return;

    try {
      await navigator.clipboard.writeText(createdKey.plain_key);
      setCopied(true);
      toast.success("API Key copied to clipboard");

      // Reset copied state after 2 seconds
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      toast.error("Failed to copy to clipboard");
      console.error("Copy failed:", error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className="glass-card hover-glow-teal">
          <Plus className="mr-2 h-4 w-4" />
          Create API Key
        </Button>
      </DialogTrigger>
      <DialogContent className="glass-card-enhanced sm:max-w-2xl">
        {step === "form" ? (
          <>
            <DialogHeader>
              <DialogTitle className="text-2xl gradient-text">
                Create API Key
              </DialogTitle>
              <DialogDescription className="text-muted-foreground">
                Generate a new API key with specific permissions. The key will
                be displayed only once.
              </DialogDescription>
            </DialogHeader>

            <form
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                form.handleSubmit();
              }}
              className="space-y-6 mt-4"
            >
              <FieldGroup>
                <FieldSet>
                  <FieldLegend>API Key Configuration</FieldLegend>
                  <FieldDescription>
                    Configure the role and description for your new API key
                  </FieldDescription>

                  <FieldGroup>
                    {/* Role Selection */}
                    <form.Field
                      name="role"
                      validators={{
                        onChange: ({ value }) => {
                          if (!value) {
                            return "Role is required";
                          }
                          if (!["read", "write", "admin"].includes(value)) {
                            return "Invalid role selected";
                          }
                          return undefined;
                        },
                      }}
                    >
                      {(field) => (
                        <Field>
                          <FieldLabel htmlFor="role">Role</FieldLabel>
                          <Select
                            value={field.state.value}
                            onValueChange={(value) =>
                              field.handleChange(
                                value as "read" | "write" | "admin",
                              )
                            }
                            disabled={createAPIKey.isPending}
                          >
                            <SelectTrigger
                              id="role"
                              className="glass-card w-full"
                            >
                              <SelectValue placeholder="Select a role" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="read">
                                <span className="flex items-center gap-2">
                                  <span>👁️</span>
                                  <span>Read - View data only</span>
                                </span>
                              </SelectItem>
                              <SelectItem value="write">
                                <span className="flex items-center gap-2">
                                  <span>✏️</span>
                                  <span>
                                    Write - Send messages and modify data
                                  </span>
                                </span>
                              </SelectItem>
                              <SelectItem value="admin">
                                <span className="flex items-center gap-2">
                                  <span>🔑</span>
                                  <span>
                                    Admin - Full access including API key
                                    management
                                  </span>
                                </span>
                              </SelectItem>
                            </SelectContent>
                          </Select>
                          {field.state.meta.errors.length > 0 && (
                            <FieldDescription className="text-destructive">
                              {field.state.meta.errors[0]}
                            </FieldDescription>
                          )}
                          <FieldDescription>
                            Choose the permission level for this API key
                          </FieldDescription>
                        </Field>
                      )}
                    </form.Field>

                    {/* Description */}
                    <form.Field
                      name="description"
                      validators={{
                        onChange: ({ value }) => {
                          if (value && value.length > 500) {
                            return "Description must be less than 500 characters";
                          }
                          return undefined;
                        },
                      }}
                    >
                      {(field) => (
                        <Field>
                          <FieldLabel htmlFor="description">
                            Description (Optional)
                          </FieldLabel>
                          <Textarea
                            id="description"
                            placeholder="e.g., Production server key for sending notifications"
                            maxLength={500}
                            rows={3}
                            value={field.state.value}
                            onChange={(e) => field.handleChange(e.target.value)}
                            onBlur={field.handleBlur}
                            disabled={createAPIKey.isPending}
                            className="glass-card resize-none"
                          />
                          {field.state.meta.errors.length > 0 && (
                            <FieldDescription className="text-destructive">
                              {field.state.meta.errors[0]}
                            </FieldDescription>
                          )}
                          <FieldDescription>
                            Add a description to help identify this key later
                            (max 500 characters)
                          </FieldDescription>
                        </Field>
                      )}
                    </form.Field>
                  </FieldGroup>
                </FieldSet>

                {/* Action Buttons */}
                <Field orientation="horizontal" className="pt-4">
                  <Button
                    type="submit"
                    disabled={form.state.isSubmitting || createAPIKey.isPending}
                    className="glass-card hover-glow-teal"
                  >
                    <Key className="mr-2 h-4 w-4" />
                    {form.state.isSubmitting || createAPIKey.isPending
                      ? "Creating..."
                      : "Create API Key"}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleClose}
                    disabled={form.state.isSubmitting || createAPIKey.isPending}
                    className="glass-card"
                  >
                    Cancel
                  </Button>
                </Field>
              </FieldGroup>
            </form>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle className="text-2xl gradient-text">
                API Key Created Successfully
              </DialogTitle>
              <DialogDescription className="text-muted-foreground">
                Save this key securely. You won't be able to see it again.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-6 mt-4">
              {/* Warning Banner */}
              <div className="glass-card p-4 border-2 border-yellow-500/20 bg-yellow-500/5 rounded-lg">
                <div className="flex items-start gap-3">
                  <span className="text-2xl">⚠️</span>
                  <div className="flex-1">
                    <h4 className="font-semibold text-yellow-600 dark:text-yellow-400 mb-1">
                      Important: Save Your API Key
                    </h4>
                    <p className="text-sm text-muted-foreground">
                      This is the only time you'll see the full API key. Make
                      sure to copy it and store it securely. If you lose it,
                      you'll need to create a new one.
                    </p>
                  </div>
                </div>
              </div>

              {/* API Key Display */}
              <FieldSet>
                <FieldLegend>Your API Key</FieldLegend>
                <div className="glass-card p-4 rounded-lg border-2 border-primary/20">
                  <div className="flex items-center gap-3">
                    <code className="flex-1 font-mono text-sm break-all bg-muted/50 p-3 rounded">
                      {createdKey?.plain_key}
                    </code>
                    <Button
                      onClick={handleCopyKey}
                      variant="outline"
                      size="icon"
                      className="glass-card shrink-0"
                      title="Copy to clipboard"
                    >
                      {copied ? (
                        <Check className="h-4 w-4 text-green-500" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
              </FieldSet>

              {/* Key Details */}
              <FieldSet>
                <FieldLegend>Key Details</FieldLegend>
                <div className="glass-card p-4 rounded-lg space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">
                      Key ID:
                    </span>
                    <code className="text-sm font-mono">
                      {createdKey?.api_key.id}
                    </code>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">Role:</span>
                    <span className="text-sm font-medium capitalize">
                      {createdKey?.api_key.role}
                    </span>
                  </div>
                  {createdKey?.api_key.description && (
                    <div className="flex justify-between items-start gap-4">
                      <span className="text-sm text-muted-foreground">
                        Description:
                      </span>
                      <span className="text-sm text-right flex-1">
                        {createdKey.api_key.description}
                      </span>
                    </div>
                  )}
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">
                      Created:
                    </span>
                    <span className="text-sm">
                      {new Date(
                        createdKey?.api_key.created_at || "",
                      ).toLocaleString()}
                    </span>
                  </div>
                </div>
              </FieldSet>

              {/* Close Button */}
              <Field orientation="horizontal" className="pt-4">
                <Button
                  onClick={handleClose}
                  className="glass-card hover-glow-teal w-full"
                >
                  Done
                </Button>
              </Field>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
