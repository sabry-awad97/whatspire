import {
  HeadContent,
  Outlet,
  createRootRouteWithContext,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

import { ErrorBoundary } from "@/components/error-boundary";
import { ErrorFallback } from "@/components/error-fallback";
import { Sidebar } from "@/components/sidebar";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";

import "../index.css";

import type { QueryClient } from "@tanstack/react-query";

export interface RouterAppContext {
  queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<RouterAppContext>()({
  component: RootComponent,
  head: () => ({
    meta: [
      {
        title: "whatspire",
      },
      {
        name: "description",
        content: "whatspire is a web application",
      },
    ],
    links: [
      {
        rel: "icon",
        href: "/favicon.ico",
      },
    ],
  }),
});

function RootComponent() {
  return (
    <>
      <HeadContent />
      <ThemeProvider
        attribute="class"
        defaultTheme="dark"
        disableTransitionOnChange
        storageKey="vite-ui-theme"
      >
        <ErrorBoundary
          fallback={<ErrorFallback />}
          onError={(error, errorInfo) => {
            // Log to console in development
            if (import.meta.env.DEV) {
              console.error("Global ErrorBoundary caught:", error);
              console.error("Component stack:", errorInfo.componentStack);
            }
            // TODO: Add production error logging (e.g., Sentry)
          }}
        >
          <div className="flex h-screen bg-background">
            {/* Glassmorphic Sidebar */}
            <Sidebar />

            {/* Main Content Area */}
            <main className="flex-1 ml-64 overflow-auto">
              <Outlet />
            </main>
          </div>
        </ErrorBoundary>
        <Toaster richColors />
      </ThemeProvider>
      <TanStackRouterDevtools position="bottom-left" />
    </>
  );
}
