import { Key, Server, Shield } from "lucide-react";
import { useTheme } from "next-themes";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "../ui/button";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel
} from "../ui/field";
import { Input } from "../ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { APIKeySettings } from "./api-key-settings";

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
    <Tabs defaultValue="general" className="space-y-6">
      <TabsList className="glass-card-enhanced">
        <TabsTrigger value="general" className="data-[state=active]:glass-card">
          <Server className="mr-2 h-4 w-4" />
          General
        </TabsTrigger>
        <TabsTrigger
          value="api-keys"
          className="data-[state=active]:glass-card"
        >
          <Shield className="mr-2 h-4 w-4" />
          API Keys
        </TabsTrigger>
      </TabsList>

      {/* General Tab */}
      <TabsContent value="general" className="space-y-6">
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

        {/* About */}
        <div className="glass-card-enhanced p-6">
          <h2 className="text-xl font-semibold mb-4">About</h2>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Version</span>
              <span className="font-medium">2.0.0</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Build</span>
              <span className="font-medium">2026.02.04</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Platform</span>
              <span className="font-medium">Desktop (Tauri)</span>
            </div>
          </div>
        </div>
      </TabsContent>

      {/* API Keys Tab */}
      <TabsContent value="api-keys">
        <APIKeySettings />
      </TabsContent>      
    </Tabs>
  );
}
