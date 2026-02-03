import { useEffect, useState } from "react";
import { Wifi, WifiOff } from "lucide-react";

import { cn } from "@/lib/utils";

// ============================================================================
// Types
// ============================================================================

type ConnectionStatus = "connected" | "connecting" | "disconnected";

interface ConnectionStatusProps {
  className?: string;
}

// ============================================================================
// Component
// ============================================================================

export function ConnectionStatus({ className }: ConnectionStatusProps) {
  const [status, setStatus] = useState<ConnectionStatus>("connected");
  const [isOnline, setIsOnline] = useState(navigator.onLine);

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      setStatus("connected");
    };

    const handleOffline = () => {
      setIsOnline(false);
      setStatus("disconnected");
    };

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  const statusConfig = {
    connected: {
      label: "Connected",
      color: "text-emerald",
      bgColor: "bg-emerald/10",
      icon: Wifi,
      pulse: false,
    },
    connecting: {
      label: "Connecting...",
      color: "text-amber",
      bgColor: "bg-amber/10",
      icon: Wifi,
      pulse: true,
    },
    disconnected: {
      label: "Disconnected",
      color: "text-destructive",
      bgColor: "bg-destructive/10",
      icon: WifiOff,
      pulse: false,
    },
  };

  const config = statusConfig[status];
  const Icon = config.icon;

  return (
    <div
      className={cn(
        "inline-flex items-center gap-2 px-3 py-1.5 rounded-lg glass-card",
        config.bgColor,
        className,
      )}
    >
      <div className="relative">
        <Icon className={cn("h-4 w-4", config.color)} />
        {config.pulse && (
          <span className="absolute inset-0 animate-ping">
            <Icon className={cn("h-4 w-4", config.color)} />
          </span>
        )}
      </div>
      <span className={cn("text-sm font-medium", config.color)}>
        {config.label}
      </span>
    </div>
  );
}
