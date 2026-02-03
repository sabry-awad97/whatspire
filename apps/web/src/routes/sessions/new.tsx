import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useForm } from "@tanstack/react-form";
import { z } from "zod";
import { ArrowLeft, Plus } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useSessionStore } from "@/stores/session-store";
import { apiClient } from "@/lib/api-client";

export const Route = createFileRoute("/sessions/new")({
  component: NewSessionPage,
});

// ============================================================================
// Component
// ============================================================================

function NewSessionPage() {
  const navigate = useNavigate();
  const { addSession } = useSessionStore();

  const form = useForm({
    defaultValues: {
      sessionName: "",
      phoneNumber: "",
      accountProtection: true,
      messageLogging: true,
      readMessages: false,
      autoRejectCalls: false,
      alwaysOnline: false,
      ignoreGroups: false,
      ignoreBroadcasts: false,
      ignoreChannels: false,
      enableWebhook: false,
      webhookUrl: "",
    },
    onSubmit: async ({ value }) => {
      try {
        // Create session via API
        const session = await apiClient.createSession({
          session_id: value.sessionName,
          name: value.sessionName,
        });

        // Add to local store
        addSession(session);
        toast.success("Session created successfully");

        // Navigate to the session details page to scan QR code
        navigate({
          to: "/sessions/$sessionId",
          params: { sessionId: session.id },
        });
      } catch (error) {
        const message =
          error instanceof Error ? error.message : "Failed to create session";
        toast.error(message);
      }
    },
  });

  return (
    <div className="min-h-screen network-bg">
      {/* Header */}
      <div className="glass-card border-b border-border/50 px-6 py-4">
        <div className="max-w-4xl mx-auto flex items-center gap-4">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => navigate({ to: "/sessions" })}
            className="glass-card hover-lift"
          >
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">Create New Session</h1>
            <p className="text-sm text-muted-foreground">
              Configure your WhatsApp session settings
            </p>
          </div>
        </div>
      </div>

      {/* Form */}
      <div className="max-w-4xl mx-auto p-6">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
          <div className="glass-card-enhanced p-6 space-y-6">
            <div>
              <h2 className="text-xl font-semibold mb-2">
                Create WhatsApp Session
              </h2>
              <p className="text-sm text-muted-foreground">
                Configure your WhatsApp session. You'll scan a QR code after
                creation.
              </p>
            </div>

            <FieldGroup>
              {/* Session Name */}
              <form.Field
                name="sessionName"
                validators={{
                  onChange: ({ value }) => {
                    if (!value || value.length === 0) {
                      return "Session name is required";
                    }
                    if (value.length > 50) {
                      return "Session name must be less than 50 characters";
                    }
                    return undefined;
                  },
                }}
              >
                {(field) => (
                  <Field>
                    <FieldLabel htmlFor="sessionName">
                      Session Name{" "}
                      <span className="text-xs text-muted-foreground">
                        (used to identify different WhatsApp sessions)
                      </span>
                    </FieldLabel>
                    <Input
                      id="sessionName"
                      placeholder="e.g., business-account"
                      value={field.state.value}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                      className="glass-card"
                    />
                    {field.state.meta.errors.length > 0 && (
                      <FieldDescription className="text-destructive">
                        {field.state.meta.errors[0]}
                      </FieldDescription>
                    )}
                  </Field>
                )}
              </form.Field>

              {/* Phone Number */}
              <form.Field
                name="phoneNumber"
                validators={{
                  onChange: ({ value }) => {
                    if (!value || value === "") return undefined;
                    if (!/^\+?[1-9]\d{9,14}$/.test(value)) {
                      return "Invalid phone number format";
                    }
                    return undefined;
                  },
                }}
              >
                {(field) => (
                  <Field>
                    <FieldLabel htmlFor="phoneNumber">Phone Number</FieldLabel>
                    <Input
                      id="phoneNumber"
                      type="tel"
                      placeholder="+1234567890"
                      value={field.state.value}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                      className="glass-card"
                    />
                    {field.state.meta.errors.length > 0 && (
                      <FieldDescription className="text-destructive">
                        {field.state.meta.errors[0]}
                      </FieldDescription>
                    )}
                  </Field>
                )}
              </form.Field>

              {/* Account Protection */}
              <form.Field name="accountProtection">
                {(field) => (
                  <Field orientation="horizontal">
                    <Checkbox
                      id="accountProtection"
                      checked={field.state.value}
                      onCheckedChange={(checked) =>
                        field.handleChange(checked as boolean)
                      }
                    />
                    <div className="flex-1">
                      <FieldLabel
                        htmlFor="accountProtection"
                        className="font-normal"
                      >
                        Enable Account Protection
                      </FieldLabel>
                      <FieldDescription>
                        Helps prevent WhatsApp from restricting your account by
                        controlling message sending frequency.
                      </FieldDescription>
                    </div>
                  </Field>
                )}
              </form.Field>

              {/* Message Logging */}
              <form.Field name="messageLogging">
                {(field) => (
                  <Field orientation="horizontal">
                    <Checkbox
                      id="messageLogging"
                      checked={field.state.value}
                      onCheckedChange={(checked) =>
                        field.handleChange(checked as boolean)
                      }
                    />
                    <div className="flex-1">
                      <FieldLabel
                        htmlFor="messageLogging"
                        className="font-normal"
                      >
                        Enable Message Logging
                      </FieldLabel>
                      <FieldDescription>
                        When disabled, only delivery statuses are recorded. When
                        enabled, full message content and recipient details are
                        stored.
                      </FieldDescription>
                    </div>
                  </Field>
                )}
              </form.Field>

              {/* Read Incoming Messages */}
              <form.Field name="readMessages">
                {(field) => (
                  <Field orientation="horizontal">
                    <Checkbox
                      id="readMessages"
                      checked={field.state.value}
                      onCheckedChange={(checked) =>
                        field.handleChange(checked as boolean)
                      }
                    />
                    <div className="flex-1">
                      <FieldLabel
                        htmlFor="readMessages"
                        className="font-normal"
                      >
                        Read Incoming Messages
                      </FieldLabel>
                      <FieldDescription>
                        When enabled, messages will be marked as read
                        automatically when received. This lets senders know
                        you've seen their messages.
                      </FieldDescription>
                    </div>
                  </Field>
                )}
              </form.Field>

              {/* Auto Reject Calls */}
              <form.Field name="autoRejectCalls">
                {(field) => (
                  <Field orientation="horizontal">
                    <Checkbox
                      id="autoRejectCalls"
                      checked={field.state.value}
                      onCheckedChange={(checked) =>
                        field.handleChange(checked as boolean)
                      }
                    />
                    <div className="flex-1">
                      <FieldLabel
                        htmlFor="autoRejectCalls"
                        className="font-normal"
                      >
                        Auto Reject Calls
                      </FieldLabel>
                      <FieldDescription>
                        When enabled, incoming calls will be automatically
                        rejected.
                      </FieldDescription>
                    </div>
                  </Field>
                )}
              </form.Field>

              {/* Always Online */}
              <form.Field name="alwaysOnline">
                {(field) => (
                  <Field orientation="horizontal">
                    <Checkbox
                      id="alwaysOnline"
                      checked={field.state.value}
                      onCheckedChange={(checked) =>
                        field.handleChange(checked as boolean)
                      }
                    />
                    <div className="flex-1">
                      <FieldLabel
                        htmlFor="alwaysOnline"
                        className="font-normal"
                      >
                        Always Online
                      </FieldLabel>
                      <FieldDescription>
                        When enabled, your session will always appear online to
                        your contacts, even when you're not actively using
                        WhatsApp. This can be useful for business accounts to
                        let customers know you're available.
                      </FieldDescription>
                    </div>
                  </Field>
                )}
              </form.Field>

              {/* Message Filtering */}
              <Collapsible className="glass-card p-4 rounded-lg">
                <CollapsibleTrigger className="flex items-center justify-between w-full">
                  <div className="text-left">
                    <h3 className="font-medium">Message Filtering</h3>
                    <p className="text-sm text-muted-foreground">
                      Choose which message types to ignore
                    </p>
                  </div>
                  <svg
                    className="h-5 w-5 transition-transform"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </CollapsibleTrigger>
                <CollapsibleContent className="pt-4 space-y-4">
                  <form.Field name="ignoreGroups">
                    {(field) => (
                      <Field orientation="horizontal">
                        <Checkbox
                          id="ignoreGroups"
                          checked={field.state.value}
                          onCheckedChange={(checked) =>
                            field.handleChange(checked as boolean)
                          }
                        />
                        <FieldLabel
                          htmlFor="ignoreGroups"
                          className="font-normal"
                        >
                          Ignore Groups
                        </FieldLabel>
                      </Field>
                    )}
                  </form.Field>

                  <form.Field name="ignoreBroadcasts">
                    {(field) => (
                      <Field orientation="horizontal">
                        <Checkbox
                          id="ignoreBroadcasts"
                          checked={field.state.value}
                          onCheckedChange={(checked) =>
                            field.handleChange(checked as boolean)
                          }
                        />
                        <FieldLabel
                          htmlFor="ignoreBroadcasts"
                          className="font-normal"
                        >
                          Ignore Broadcasts
                        </FieldLabel>
                      </Field>
                    )}
                  </form.Field>

                  <form.Field name="ignoreChannels">
                    {(field) => (
                      <Field orientation="horizontal">
                        <Checkbox
                          id="ignoreChannels"
                          checked={field.state.value}
                          onCheckedChange={(checked) =>
                            field.handleChange(checked as boolean)
                          }
                        />
                        <FieldLabel
                          htmlFor="ignoreChannels"
                          className="font-normal"
                        >
                          Ignore Channels
                        </FieldLabel>
                      </Field>
                    )}
                  </form.Field>
                </CollapsibleContent>
              </Collapsible>

              {/* Webhook Notifications */}
              <form.Field name="enableWebhook">
                {(enableWebhookField) => (
                  <FieldSet className="glass-card p-4 rounded-lg">
                    <Field orientation="horizontal">
                      <Checkbox
                        id="enableWebhook"
                        checked={enableWebhookField.state.value}
                        onCheckedChange={(checked) =>
                          enableWebhookField.handleChange(checked as boolean)
                        }
                      />
                      <div className="flex-1">
                        <FieldLabel
                          htmlFor="enableWebhook"
                          className="font-normal"
                        >
                          Enable Webhook Notifications (Optional)
                        </FieldLabel>
                        <FieldDescription>
                          When enabled, events will be sent to the webhook URL
                          above.
                        </FieldDescription>
                      </div>
                    </Field>

                    {enableWebhookField.state.value && (
                      <form.Field
                        name="webhookUrl"
                        validators={{
                          onChange: ({ value }) => {
                            if (!value || value === "") return undefined;
                            try {
                              new URL(value);
                              return undefined;
                            } catch {
                              return "Invalid URL";
                            }
                          },
                        }}
                      >
                        {(field) => (
                          <Field className="mt-4">
                            <FieldLabel htmlFor="webhookUrl">
                              Webhook URL
                            </FieldLabel>
                            <Input
                              id="webhookUrl"
                              type="url"
                              placeholder="https://your-domain.com/webhook"
                              value={field.state.value}
                              onChange={(e) =>
                                field.handleChange(e.target.value)
                              }
                              onBlur={field.handleBlur}
                              className="glass-card"
                            />
                            {field.state.meta.errors.length > 0 && (
                              <FieldDescription className="text-destructive">
                                {field.state.meta.errors[0]}
                              </FieldDescription>
                            )}
                          </Field>
                        )}
                      </form.Field>
                    )}
                  </FieldSet>
                )}
              </form.Field>
            </FieldGroup>

            {/* Action Buttons */}
            <div className="flex items-center justify-end gap-3 pt-4 border-t border-border/50">
              <Button
                type="button"
                variant="outline"
                onClick={() => navigate({ to: "/sessions" })}
                disabled={form.state.isSubmitting}
                className="glass-card hover-glow-teal"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={form.state.isSubmitting}
                className="glass-card hover-glow-emerald"
              >
                <Plus className="mr-2 h-4 w-4" />
                {form.state.isSubmitting ? "Creating..." : "Create Session"}
              </Button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
