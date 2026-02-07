import { createFileRoute, Link } from "@tanstack/react-router";
import { useSessions, useSessionEvents } from "@/hooks";
import { Activity, MessageSquare, TrendingUp, Users, Zap } from "lucide-react";

export const Route = createFileRoute("/")({
  component: HomeComponent,
});

function HomeComponent() {
  const { data: sessions = [] } = useSessions();

  // Enable real-time session status updates
  useSessionEvents();

  // Calculate stats
  const connectedSessions = sessions.filter(
    (s) => s.status === "connected",
  ).length;
  const totalSessions = sessions.length;
  const pendingSessions = sessions.filter((s) => s.status === "pending").length;

  return (
    <div className="min-h-screen network-bg p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <div className="space-y-2">
          <h1 className="text-3xl font-bold gradient-text">
            Dashboard Overview
          </h1>
          <p className="text-muted-foreground">
            Welcome back. Here's your WhatsApp service summary.
          </p>
        </div>

        {/* Status Badge */}
        <div className="flex items-center gap-2">
          <div className="glass-card px-4 py-2 inline-flex items-center gap-2 glow-emerald">
            <div className="w-2 h-2 rounded-full bg-emerald pulse-live" />
            <span className="text-sm font-medium text-emerald">
              All Systems Operational
            </span>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Total Sessions */}
          <div className="glass-card-enhanced p-6 hover-lift hover-glow-teal animate-scale-in">
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 rounded-xl glass-card glow-teal-sm bg-teal/10">
                <Zap className="w-6 h-6 text-teal" />
              </div>
              <TrendingUp className="w-5 h-5 text-emerald" />
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Total Sessions</p>
              <p className="text-3xl font-bold text-teal">{totalSessions}</p>
              <p className="text-xs text-muted-foreground">All Regions</p>
            </div>
          </div>

          {/* Connected Sessions */}
          <div
            className="glass-card-enhanced p-6 hover-lift hover-glow-emerald animate-scale-in"
            style={{ animationDelay: "0.1s" }}
          >
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 rounded-xl glass-card glow-emerald bg-emerald/10">
                <Activity className="w-6 h-6 text-emerald" />
              </div>
              <div className="text-xs font-medium text-emerald">
                {totalSessions > 0
                  ? Math.round((connectedSessions / totalSessions) * 100)
                  : 0}
                %
              </div>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Connected</p>
              <p className="text-3xl font-bold text-emerald">
                {connectedSessions}
              </p>
              <p className="text-xs text-muted-foreground">Active Trading</p>
            </div>
          </div>

          {/* Pending Sessions */}
          <div
            className="glass-card-enhanced p-6 hover-lift hover-glow-amber animate-scale-in"
            style={{ animationDelay: "0.2s" }}
          >
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 rounded-xl glass-card glow-amber bg-amber/10">
                <MessageSquare className="w-6 h-6 text-amber" />
              </div>
              <div className="text-xs font-medium text-amber">
                {pendingSessions > 0 ? "+8%" : "0%"}
              </div>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Pending Auth</p>
              <p className="text-3xl font-bold text-amber">{pendingSessions}</p>
              <p className="text-xs text-muted-foreground">Awaiting QR Scan</p>
            </div>
          </div>

          {/* System Status */}
          <div
            className="glass-card-enhanced p-6 hover-lift animate-scale-in"
            style={{ animationDelay: "0.3s" }}
          >
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 rounded-xl glass-card bg-primary/10">
                <Users className="w-6 h-6 text-primary" />
              </div>
              <div className="flex items-center gap-1">
                <div className="w-2 h-2 rounded-full bg-emerald pulse-live" />
                <span className="text-xs text-emerald">Live</span>
              </div>
            </div>
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">System Health</p>
              <p className="text-3xl font-bold text-foreground">"100%"</p>
              <p className="text-xs text-muted-foreground">Uptime</p>
            </div>
          </div>
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Recent Activity */}
          <div
            className="lg:col-span-2 glass-card-enhanced p-6 animate-scale-in"
            style={{ animationDelay: "0.4s" }}
          >
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-xl font-semibold">Recent Activity</h2>
                <p className="text-sm text-muted-foreground mt-1">
                  Live updates from your sessions
                </p>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald pulse-live" />
                <span className="text-xs text-emerald">Live</span>
              </div>
            </div>

            {/* Activity List */}
            <div className="space-y-3">
              {sessions.length === 0 ? (
                <div className="text-center py-12">
                  <MessageSquare className="w-12 h-12 text-muted-foreground mx-auto mb-3 opacity-50" />
                  <p className="text-sm text-muted-foreground">
                    No sessions yet. Create your first session to get started.
                  </p>
                </div>
              ) : (
                sessions.slice(0, 5).map((session) => (
                  <div
                    key={session.id}
                    className="glass-card p-4 hover-lift ripple flex items-center gap-4"
                  >
                    <div
                      className={`w-10 h-10 rounded-lg glass-card flex items-center justify-center ${session.status === "connected"
                        ? "glow-emerald bg-emerald/10"
                        : session.status === "pending"
                          ? "glow-amber bg-amber/10"
                          : "bg-muted/10"
                        }`}
                    >
                      <Zap
                        className={`w-5 h-5 ${session.status === "connected"
                          ? "text-emerald"
                          : session.status === "pending"
                            ? "text-amber"
                            : "text-muted-foreground"
                          }`}
                      />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">{session.id}</p>
                      <p className="text-xs text-muted-foreground">
                        {session.status === "connected"
                          ? "Connected"
                          : session.status === "pending"
                            ? "Pending Authentication"
                            : "Disconnected"}
                      </p>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(session.created_at).toLocaleDateString()}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Quick Actions */}
          <div
            className="glass-card-enhanced p-6 animate-scale-in"
            style={{ animationDelay: "0.5s" }}
          >
            <h2 className="text-xl font-semibold mb-6">Quick Actions</h2>
            <div className="space-y-3">
              <Link
                to="/sessions"
                className="block glass-card p-4 hover-lift hover-glow-teal ripple transition-all"
              >
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg glass-card glow-teal-sm bg-teal/10">
                    <Zap className="w-5 h-5 text-teal" />
                  </div>
                  <div>
                    <p className="font-medium">Manage Sessions</p>
                    <p className="text-xs text-muted-foreground">
                      View and control sessions
                    </p>
                  </div>
                </div>
              </Link>

              <Link
                to="/sessions/new"
                className="block glass-card p-4 hover-lift hover-glow-emerald ripple transition-all"
              >
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg glass-card glow-emerald bg-emerald/10">
                    <Zap className="w-5 h-5 text-emerald" />
                  </div>
                  <div>
                    <p className="font-medium">New Session</p>
                    <p className="text-xs text-muted-foreground">
                      Create a new WhatsApp session
                    </p>
                  </div>
                </div>
              </Link>

              <Link
                to="/settings"
                className="block glass-card p-4 hover-lift hover-glow-amber ripple transition-all"
              >
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg glass-card glow-amber bg-amber/10">
                    <Users className="w-5 h-5 text-amber" />
                  </div>
                  <div>
                    <p className="font-medium">Settings</p>
                    <p className="text-xs text-muted-foreground">
                      Manage API keys and preferences
                    </p>
                  </div>
                </div>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
