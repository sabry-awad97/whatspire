import { Bell, Key, Moon, Server, Sun } from "lucide-react";
import { useState } from "react";
import { useTheme } from "next-themes";
import { toast } from "sonner";

import { Button } from "../ui/button";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "../ui/field";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Separator } from "../ui/separator";
import { Switch } from "../ui/switch";

// ============================================================================
// Component
// ============================================================================

export function Settings() {
  const { theme, setTheme } = useTheme();
  const [apiEndpoint, setApiEndpoint] = useState("https://api.whatspire.com");
  const [apiKey, setApiKey] = useState("");
  const [showApiKey, setShowApiKey] = useState(false);

  // Notification preferences
  const [notifications, setNotifications] = useState({
    newMessages: true,
    sessionStatus: true,
    systemAlerts: false,
    soundEnabled: true,
    desktopNotifications: true,
  });

  const handleSaveApiSettings = () => {
    toast.success("API settings saved successfully");
  };

  const handleSaveNotifications = () => {
    toast.success("Notification preferences saved");
  };

  const handleTestConnection = async () => {
    toast.info("Testing connection...");
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1500));
    toast.success("Connection successful!");
  };

  return (
    <div className="space-y-6">
      {/* API Configuration */}
      <div className="glass-card-enhanced p-6">
        <div className="flex items-center gap-3 mb-6">
          <div className="p-2 rounded-lg glass-card bg-teal/10">
            <Server className="h-5 w-5 text-teal" />
          </div>
          <div>
            <h2 className="text-xl font-semibold">API Configuration</h2>
            <p className="text-sm text-muted-foreground">
              Configure your API endpoint and authentication
            </p>
          </div>
        </div>

        <FieldGroup>
          {/* API Endpoint */}
          <Field>
            <FieldLabel htmlFor="api-endpoint">API Endpoint</FieldLabel>
            <Input
              id="api-endpoint"
              type="url"
              value={apiEndpoint}
              onChange={(e) => setApiEndpoint(e.target.value)}
              placeholder="https://api.whatspire.com"
              className="glass-card"
            />
            <FieldDescription>
              The base URL for the Whatspire API server
            </FieldDescription>
          </Field>

          {/* API Key */}
          <Field>
            <FieldLabel htmlFor="api-key">API Key</FieldLabel>
            <div className="flex gap-2">
              <Input
                id="api-key"
                type={showApiKey ? "text" : "password"}
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Enter your API key"
                className="glass-card flex-1"
              />
              <Button
                variant="outline"
                onClick={() => setShowApiKey(!showApiKey)}
                className="glass-card hover-glow-teal"
              >
                {showApiKey ? "Hide" : "Show"}
              </Button>
            </div>
            <FieldDescription>
              Your authentication key for API requests
            </FieldDescription>
          </Field>

          {/* Action Buttons */}
          <div className="flex items-center gap-3 pt-4">
            <Button
              onClick={handleTestConnection}
              variant="outline"
              className="glass-card hover-glow-emerald"
            >
              Test Connection
            </Button>
            <Button
              onClick={handleSaveApiSettings}
              className="glass-card hover-glow-teal"
            >
              <Key className="mr-2 h-4 w-4" />
              Save API Settings
            </Button>
          </div>
        </FieldGroup>
      </div>

      {/* Appearance */}
      <div className="glass-card-enhanced p-6">
        <div className="flex items-center gap-3 mb-6">
          <div className="p-2 rounded-lg glass-card bg-amber/10">
            {theme === "dark" ? (
              <Moon className="h-5 w-5 text-amber" />
            ) : (
              <Sun className="h-5 w-5 text-amber" />
            )}
          </div>
          <div>
            <h2 className="text-xl font-semibold">Appearance</h2>
            <p className="text-sm text-muted-foreground">
              Customize the look and feel of the application
            </p>
          </div>
        </div>

        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="theme">Theme</FieldLabel>
            <Select value={theme} onValueChange={setTheme}>
              <SelectTrigger id="theme" className="glass-card">
                <SelectValue placeholder="Select theme" />
              </SelectTrigger>
              <SelectContent className="glass-card-enhanced">
                <SelectItem value="light">Light</SelectItem>
                <SelectItem value="dark">Dark</SelectItem>
                <SelectItem value="system">System</SelectItem>
              </SelectContent>
            </Select>
            <FieldDescription>
              Choose your preferred color scheme
            </FieldDescription>
          </Field>
        </FieldGroup>
      </div>

      {/* Notifications */}
      <div className="glass-card-enhanced p-6">
        <div className="flex items-center gap-3 mb-6">
          <div className="p-2 rounded-lg glass-card bg-emerald/10">
            <Bell className="h-5 w-5 text-emerald" />
          </div>
          <div>
            <h2 className="text-xl font-semibold">Notifications</h2>
            <p className="text-sm text-muted-foreground">
              Manage your notification preferences
            </p>
          </div>
        </div>

        <FieldSet>
          <FieldGroup>
            {/* New Messages */}
            <div className="flex items-center justify-between py-3">
              <div className="space-y-0.5">
                <Label htmlFor="new-messages" className="text-base">
                  New Messages
                </Label>
                <p className="text-sm text-muted-foreground">
                  Get notified when you receive new messages
                </p>
              </div>
              <Switch
                id="new-messages"
                checked={notifications.newMessages}
                onCheckedChange={(checked) =>
                  setNotifications({ ...notifications, newMessages: checked })
                }
              />
            </div>

            <Separator />

            {/* Session Status */}
            <div className="flex items-center justify-between py-3">
              <div className="space-y-0.5">
                <Label htmlFor="session-status" className="text-base">
                  Session Status Changes
                </Label>
                <p className="text-sm text-muted-foreground">
                  Get notified when sessions connect or disconnect
                </p>
              </div>
              <Switch
                id="session-status"
                checked={notifications.sessionStatus}
                onCheckedChange={(checked) =>
                  setNotifications({ ...notifications, sessionStatus: checked })
                }
              />
            </div>

            <Separator />

            {/* System Alerts */}
            <div className="flex items-center justify-between py-3">
              <div className="space-y-0.5">
                <Label htmlFor="system-alerts" className="text-base">
                  System Alerts
                </Label>
                <p className="text-sm text-muted-foreground">
                  Get notified about system updates and maintenance
                </p>
              </div>
              <Switch
                id="system-alerts"
                checked={notifications.systemAlerts}
                onCheckedChange={(checked) =>
                  setNotifications({ ...notifications, systemAlerts: checked })
                }
              />
            </div>

            <Separator />

            {/* Sound */}
            <div className="flex items-center justify-between py-3">
              <div className="space-y-0.5">
                <Label htmlFor="sound-enabled" className="text-base">
                  Sound Notifications
                </Label>
                <p className="text-sm text-muted-foreground">
                  Play sound when receiving notifications
                </p>
              </div>
              <Switch
                id="sound-enabled"
                checked={notifications.soundEnabled}
                onCheckedChange={(checked) =>
                  setNotifications({ ...notifications, soundEnabled: checked })
                }
              />
            </div>

            <Separator />

            {/* Desktop Notifications */}
            <div className="flex items-center justify-between py-3">
              <div className="space-y-0.5">
                <Label htmlFor="desktop-notifications" className="text-base">
                  Desktop Notifications
                </Label>
                <p className="text-sm text-muted-foreground">
                  Show notifications on your desktop
                </p>
              </div>
              <Switch
                id="desktop-notifications"
                checked={notifications.desktopNotifications}
                onCheckedChange={(checked) =>
                  setNotifications({
                    ...notifications,
                    desktopNotifications: checked,
                  })
                }
              />
            </div>
          </FieldGroup>

          <div className="pt-4">
            <Button
              onClick={handleSaveNotifications}
              className="glass-card hover-glow-teal"
            >
              Save Notification Preferences
            </Button>
          </div>
        </FieldSet>
      </div>

      {/* About */}
      <div className="glass-card-enhanced p-6">
        <h2 className="text-xl font-semibold mb-4">About</h2>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Version</span>
            <span className="font-medium">1.0.0</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Build</span>
            <span className="font-medium">2026.02.03</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Platform</span>
            <span className="font-medium">Desktop (Tauri)</span>
          </div>
        </div>
      </div>
    </div>
  );
}
