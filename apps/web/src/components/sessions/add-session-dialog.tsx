import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowLeft, Plus, QrCode } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import type { Session } from "@/lib/api-client";
import { useSessionStore } from "@/stores/session-store";

import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
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
  FieldSeparator,
  FieldSet,
} from "../ui/field";
import { Input } from "../ui/input";
import { QRCodeDisplay } from "./qr-code-display";

// ============================================================================
// Validation Schema
// ============================================================================

const sessionConfigSchema = z
  .object({
    sessionName: z
      .string()
      .min(1, "Session name is required")
      .max(50, "Session name must be less than 50 characters"),
    phoneNumber: z
      .string()
      .optional()
      .refine(
        (val) => !val || /^\+?\d{10,15}$/.test(val.replace(/\s/g, "")),
        "Please enter a valid phone number (10-15 digits)",
      ),
    enableAccountProtection: z.boolean(),
    enableMessageLogging: z.boolean(),
    readIncomingMessages: z.boolean(),
    autoRejectCalls: z.boolean(),
    alwaysOnline: z.boolean(),
    enableWebhookNotifications: z.boolean(),
    webhookUrl: z.string().optional(),
    ignoreGroups: z.boolean(),
    ignoreBroadcasts: z.boolean(),
    ignoreChannels: z.boolean(),
  })
  .refine(
    (data) => {
      if (data.enableWebhookNotifications && !data.webhookUrl) {
        return false;
      }
      if (data.webhookUrl) {
        try {
          new URL(data.webhookUrl);
          return true;
        } catch {
          return false;
        }
      }
      return true;
    },
    {
      message: "Please enter a valid webhook URL",
      path: ["webhookUrl"],
    },
  );

type SessionConfigFormData = z.infer<typeof sessionConfigSchema>;

// ============================================================================
// Types
// ============================================================================

type DialogStep = "form" | "qr";

// ============================================================================
// Component
// ============================================================================

export function AddSessionDialog() {
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState<DialogStep>("form");
  const [isLoading, setIsLoading] = useState(false);
  const [showMessageFiltering, setShowMessageFiltering] = useState(false);
  const [sessionId, setSessionId] = useState("");
  const { addSession } = useSessionStore();

  const form = useForm<SessionConfigFormData>({
    resolver: zodResolver(sessionConfigSchema),
    defaultValues: {
      sessionName: "",
      phoneNumber: "",
      enableAccountProtection: true,
      enableMessageLogging: true,
      readIncomingMessages: false,
      autoRejectCalls: false,
      alwaysOnline: true,
      enableWebhookNotifications: false,
      webhookUrl: "",
      ignoreGroups: false,
      ignoreBroadcasts: false,
      ignoreChannels: false,
    },
  });

  const watchWebhookEnabled = form.watch("enableWebhookNotifications");

  const handleReset = () => {
    setStep("form");
    setSessionId("");
    form.reset();
    setIsLoading(false);
    setShowMessageFiltering(false);
  };

  const handleClose = () => {
    setOpen(false);
    setTimeout(handleReset, 300);
  };

  const onSubmit = async (data: SessionConfigFormData) => {
    // Generate session ID from name
    const generatedSessionId = data.sessionName
      .toLowerCase()
      .replace(/\s+/g, "-");

    setIsLoading(true);
    try {
      // Mock session creation - just add to store
      await new Promise((resolve) => setTimeout(resolve, 1000));

      const newSession: Session = {
        id: generatedSessionId,
        status: "pending",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };

      addSession(newSession);
      toast.success("Session created successfully");

      // Store session ID and move to QR step
      setSessionId(generatedSessionId);
      setStep("qr");
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Failed to create session",
      );
    } finally {
      setIsLoading(false);
    }
  };

  const handleAuthenticated = () => {
    toast.success("Session authenticated!");
    setTimeout(handleClose, 2000);
  };

  const handleError = (message: string) => {
    console.error("QR authentication error:", message);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className="glass-card hover-glow-teal">
          <Plus className="mr-2 h-4 w-4" />
          Add Session
        </Button>
      </DialogTrigger>
      <DialogContent className="glass-card-enhanced sm:max-w-2xl max-h-[90vh] overflow-y-auto">
        {step === "form" ? (
          <>
            <DialogHeader>
              <DialogTitle className="text-2xl gradient-text">
                Create WhatsApp Session
              </DialogTitle>
              <DialogDescription className="text-muted-foreground">
                Set up a new WhatsApp session. You will need to scan a QR code
                to connect after creating.
              </DialogDescription>
            </DialogHeader>

            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="space-y-6 mt-4"
            >
              <FieldGroup>
                {/* Link Your WhatsApp Number Section */}
                <FieldSet>
                  <FieldLegend>Link Your WhatsApp Number</FieldLegend>
                  <FieldDescription>
                    Set up a new WhatsApp session. You will need to scan a QR
                    code to connect after creating.
                  </FieldDescription>

                  <FieldGroup>
                    {/* Session Name */}
                    <Field>
                      <FieldLabel htmlFor="session-name">
                        Session Name
                      </FieldLabel>
                      <Input
                        id="session-name"
                        placeholder="My WhatsApp Session"
                        {...form.register("sessionName")}
                        disabled={isLoading}
                        className="glass-card"
                      />
                      {form.formState.errors.sessionName && (
                        <p className="text-sm text-destructive">
                          {form.formState.errors.sessionName.message}
                        </p>
                      )}
                      <FieldDescription>
                        Used to identify different WhatsApp sessions
                      </FieldDescription>
                    </Field>

                    {/* Phone Number */}
                    <Field>
                      <FieldLabel htmlFor="phone-number">
                        Phone Number
                      </FieldLabel>
                      <div className="flex gap-2">
                        <div className="glass-card px-3 py-2 flex items-center gap-2 rounded-lg">
                          <span className="text-sm">📱</span>
                          <span className="text-sm text-muted-foreground">
                            +
                          </span>
                        </div>
                        <Input
                          id="phone-number"
                          type="tel"
                          placeholder="1234567890"
                          {...form.register("phoneNumber")}
                          disabled={isLoading}
                          className="glass-card flex-1"
                        />
                      </div>
                      {form.formState.errors.phoneNumber && (
                        <p className="text-sm text-destructive">
                          {form.formState.errors.phoneNumber.message}
                        </p>
                      )}
                      <FieldDescription>
                        Optional - Enter the phone number you'll scan the QR
                        code with
                      </FieldDescription>
                    </Field>
                  </FieldGroup>
                </FieldSet>

                <FieldSeparator />

                {/* Configuration Options */}
                <FieldSet>
                  <FieldLegend>Session Configuration</FieldLegend>
                  <FieldDescription>
                    Configure how your WhatsApp session behaves
                  </FieldDescription>

                  <FieldGroup>
                    {/* Enable Account Protection */}
                    <Field orientation="horizontal">
                      <Checkbox
                        id="account-protection"
                        checked={form.watch("enableAccountProtection")}
                        onCheckedChange={(checked) =>
                          form.setValue(
                            "enableAccountProtection",
                            checked as boolean,
                          )
                        }
                      />
                      <div className="flex-1">
                        <FieldLabel
                          htmlFor="account-protection"
                          className="font-medium cursor-pointer"
                        >
                          Enable Account Protection
                        </FieldLabel>
                        <FieldDescription>
                          Helps prevent WhatsApp from restricting your account
                          by controlling message sending frequency
                        </FieldDescription>
                      </div>
                    </Field>

                    {/* Enable Message Logging */}
                    <Field orientation="horizontal">
                      <Checkbox
                        id="message-logging"
                        checked={form.watch("enableMessageLogging")}
                        onCheckedChange={(checked) =>
                          form.setValue(
                            "enableMessageLogging",
                            checked as boolean,
                          )
                        }
                      />
                      <div className="flex-1">
                        <FieldLabel
                          htmlFor="message-logging"
                          className="font-medium cursor-pointer"
                        >
                          Enable Message Logging
                        </FieldLabel>
                        <FieldDescription>
                          When disabled, only delivery statuses are recorded.
                          When enabled, full message content and recipient
                          details are stored
                        </FieldDescription>
                      </div>
                    </Field>

                    {/* Read Incoming Messages */}
                    <Field orientation="horizontal" className="ml-6">
                      <Checkbox
                        id="read-messages"
                        checked={form.watch("readIncomingMessages")}
                        onCheckedChange={(checked) =>
                          form.setValue(
                            "readIncomingMessages",
                            checked as boolean,
                          )
                        }
                      />
                      <div className="flex-1">
                        <FieldLabel
                          htmlFor="read-messages"
                          className="font-medium cursor-pointer"
                        >
                          Read Incoming Messages
                        </FieldLabel>
                        <FieldDescription>
                          When enabled, messages will be marked as read
                          automatically when received
                        </FieldDescription>
                      </div>
                    </Field>

                    {/* Auto Reject Calls */}
                    <Field orientation="horizontal" className="ml-6">
                      <Checkbox
                        id="reject-calls"
                        checked={form.watch("autoRejectCalls")}
                        onCheckedChange={(checked) =>
                          form.setValue("autoRejectCalls", checked as boolean)
                        }
                      />
                      <div className="flex-1">
                        <FieldLabel
                          htmlFor="reject-calls"
                          className="font-medium cursor-pointer"
                        >
                          Auto Reject Calls
                        </FieldLabel>
                        <FieldDescription>
                          When enabled, incoming calls will be automatically
                          rejected
                        </FieldDescription>
                      </div>
                    </Field>

                    {/* Always Online */}
                    <Field orientation="horizontal">
                      <Checkbox
                        id="always-online"
                        checked={form.watch("alwaysOnline")}
                        onCheckedChange={(checked) =>
                          form.setValue("alwaysOnline", checked as boolean)
                        }
                      />
                      <div className="flex-1">
                        <FieldLabel
                          htmlFor="always-online"
                          className="font-medium cursor-pointer"
                        >
                          Always Online
                        </FieldLabel>
                        <FieldDescription>
                          Your session will always appear online to your
                          contacts, even when you're not actively using WhatsApp
                        </FieldDescription>
                      </div>
                    </Field>
                  </FieldGroup>
                </FieldSet>

                <FieldSeparator />

                {/* Message Filtering (Collapsible) */}
                <FieldSet>
                  <button
                    type="button"
                    onClick={() =>
                      setShowMessageFiltering(!showMessageFiltering)
                    }
                    className="w-full flex items-center justify-between hover-lift transition-all"
                  >
                    <div className="text-left">
                      <FieldLegend>Message Filtering</FieldLegend>
                      <FieldDescription>
                        Choose which types of messages to ignore. Ignored
                        messages won't trigger webhooks or be processed
                      </FieldDescription>
                    </div>
                    <span className="text-muted-foreground ml-4 shrink-0">
                      {showMessageFiltering ? "▲" : "▼"}
                    </span>
                  </button>

                  {showMessageFiltering && (
                    <FieldGroup className="mt-4">
                      {/* Ignore Groups */}
                      <Field orientation="horizontal">
                        <Checkbox
                          id="ignore-groups"
                          checked={form.watch("ignoreGroups")}
                          onCheckedChange={(checked) =>
                            form.setValue("ignoreGroups", checked as boolean)
                          }
                        />
                        <div className="flex-1">
                          <FieldLabel
                            htmlFor="ignore-groups"
                            className="font-medium cursor-pointer"
                          >
                            Ignore Groups
                          </FieldLabel>
                          <FieldDescription>
                            Skip group messages
                          </FieldDescription>
                        </div>
                      </Field>

                      {/* Ignore Broadcasts */}
                      <Field orientation="horizontal">
                        <Checkbox
                          id="ignore-broadcasts"
                          checked={form.watch("ignoreBroadcasts")}
                          onCheckedChange={(checked) =>
                            form.setValue(
                              "ignoreBroadcasts",
                              checked as boolean,
                            )
                          }
                        />
                        <div className="flex-1">
                          <FieldLabel
                            htmlFor="ignore-broadcasts"
                            className="font-medium cursor-pointer"
                          >
                            Ignore Broadcasts
                          </FieldLabel>
                          <FieldDescription>
                            Skip broadcast lists
                          </FieldDescription>
                        </div>
                      </Field>

                      {/* Ignore Channels */}
                      <Field orientation="horizontal">
                        <Checkbox
                          id="ignore-channels"
                          checked={form.watch("ignoreChannels")}
                          onCheckedChange={(checked) =>
                            form.setValue("ignoreChannels", checked as boolean)
                          }
                        />
                        <div className="flex-1">
                          <FieldLabel
                            htmlFor="ignore-channels"
                            className="font-medium cursor-pointer"
                          >
                            Ignore Channels
                          </FieldLabel>
                          <FieldDescription>
                            Skip channel updates
                          </FieldDescription>
                        </div>
                      </Field>
                    </FieldGroup>
                  )}
                </FieldSet>

                <FieldSeparator />

                {/* Webhook Notifications */}
                <FieldSet>
                  <FieldLegend>Webhook Notifications</FieldLegend>
                  <FieldDescription>
                    Configure webhook notifications for real-time events
                  </FieldDescription>

                  <FieldGroup>
                    <Field orientation="horizontal">
                      <Checkbox
                        id="webhook-notifications"
                        checked={watchWebhookEnabled}
                        onCheckedChange={(checked) =>
                          form.setValue(
                            "enableWebhookNotifications",
                            checked as boolean,
                          )
                        }
                      />
                      <div className="flex-1">
                        <FieldLabel
                          htmlFor="webhook-notifications"
                          className="font-medium cursor-pointer"
                        >
                          Enable Webhook Notifications
                        </FieldLabel>
                        <FieldDescription>
                          When enabled, events will be sent to the webhook URL
                          below
                        </FieldDescription>
                      </div>
                    </Field>

                    {watchWebhookEnabled && (
                      <Field>
                        <FieldLabel htmlFor="webhook-url">
                          Webhook URL
                        </FieldLabel>
                        <Input
                          id="webhook-url"
                          type="url"
                          placeholder="https://your-webhook-url.com/webhook"
                          {...form.register("webhookUrl")}
                          disabled={isLoading}
                          className="glass-card"
                        />
                        {form.formState.errors.webhookUrl && (
                          <p className="text-sm text-destructive">
                            {form.formState.errors.webhookUrl.message}
                          </p>
                        )}
                        <FieldDescription>
                          Events will be sent to this URL via HTTP POST
                        </FieldDescription>
                      </Field>
                    )}
                  </FieldGroup>
                </FieldSet>

                {/* Action Buttons */}
                <Field orientation="horizontal" className="pt-4">
                  <Button
                    type="submit"
                    disabled={isLoading}
                    className="glass-card hover-glow-teal"
                  >
                    <QrCode className="mr-2 h-4 w-4" />
                    {isLoading ? "Creating..." : "Create & Connect Session"}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleClose}
                    disabled={isLoading}
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
              <div className="flex items-center gap-3">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setStep("form")}
                  className="glass-card"
                >
                  <ArrowLeft className="h-4 w-4" />
                </Button>
                <div>
                  <DialogTitle className="text-2xl gradient-text">
                    Scan QR Code
                  </DialogTitle>
                  <DialogDescription className="text-muted-foreground">
                    Scan this QR code with WhatsApp to authenticate your session
                  </DialogDescription>
                </div>
              </div>
            </DialogHeader>
            <QRCodeDisplay
              sessionId={sessionId}
              onAuthenticated={handleAuthenticated}
              onError={handleError}
            />
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
