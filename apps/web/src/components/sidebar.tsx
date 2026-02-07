import { Link, useRouterState } from "@tanstack/react-router";
import {
  Home,
  MessageSquare,
  Settings,
  Users,
  UsersRound,
  Zap,
} from "lucide-react";

import { cn } from "@/lib/utils";

// ============================================================================
// Types
// ============================================================================

interface NavItem {
  to: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}

// ============================================================================
// Navigation Items
// ============================================================================

const navItems: NavItem[] = [
  { to: "/", label: "Dashboard", icon: Home },
  { to: "/sessions", label: "Sessions", icon: Zap },
  { to: "/settings", label: "Settings", icon: Settings },
];

// ============================================================================
// Component
// ============================================================================

export function Sidebar() {
  const router = useRouterState();
  const currentPath = router.location.pathname;

  return (
    <aside className="fixed left-0 top-0 h-screen w-64 glass-card border-r border-sidebar-border">
      {/* Glassmorphic overlay with teal tint */}
      <div className="absolute inset-0 bg-gradient-to-b from-sidebar-background/90 via-sidebar-background/70 to-sidebar-background/90 backdrop-blur-2xl" />

      {/* Teal accent line */}
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-teal to-transparent opacity-60" />

      {/* Content */}
      <div className="relative h-full flex flex-col p-4">
        {/* Logo/Brand */}
        <div className="mb-8 px-2">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl glass-card glow-teal-sm flex items-center justify-center bg-teal/10">
              <Zap className="w-5 h-5 text-teal" />
            </div>
            <div>
              <h1 className="text-lg font-bold gradient-text">Whatspire</h1>
              <p className="text-xs text-muted-foreground">Broker</p>
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = currentPath === item.to;

            return (
              <Link
                key={item.to}
                to={item.to}
                className={cn(
                  "flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-300 group ripple",
                  "hover-lift",
                  isActive
                    ? "glass-card hover-glow-teal text-sidebar-accent-foreground"
                    : "text-sidebar-foreground hover:text-sidebar-accent-foreground hover:glass-card",
                )}
              >
                <Icon
                  className={cn(
                    "w-5 h-5 transition-all duration-300",
                    isActive ? "text-teal" : "group-hover:text-teal",
                  )}
                />
                <span
                  className={cn("font-medium text-sm", isActive && "text-teal")}
                >
                  {item.label}
                </span>

                {/* Active indicator */}
                {isActive && (
                  <div className="ml-auto w-1.5 h-1.5 rounded-full bg-teal glow-teal-sm animate-pulse" />
                )}
              </Link>
            );
          })}
        </nav>

        {/* Bottom Section */}
        <div className="pt-4 border-t border-sidebar-border/50 space-y-3">
          {/* Status Indicator */}
          <div className="flex items-center gap-2 px-3 py-2 rounded-lg glass-card glow-emerald">
            <div className="w-2 h-2 rounded-full bg-emerald pulse-live" />
            <span className="text-xs text-sidebar-foreground">
              All Systems Operational
            </span>
          </div>
        </div>
      </div>

      {/* Bottom teal accent line */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-teal to-transparent opacity-60" />
    </aside>
  );
}
