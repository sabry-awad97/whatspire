import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider, createRouter } from "@tanstack/react-router";
import ReactDOM from "react-dom/client";
import { WhatspireProvider } from "@whatspire/hooks";

import { ErrorFallback } from "./components/error-fallback";
import Loader from "./components/loader";
import { routeTree } from "./routeTree.gen";

// Create QueryClient with default options
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10, // 10 minutes (formerly cacheTime)
      retry: 3,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 1,
    },
  },
});

const router = createRouter({
  routeTree,
  defaultPreload: "intent",
  defaultPendingComponent: () => <Loader />,
  defaultErrorComponent: ({ error }) => (
    <ErrorFallback error={error} onReset={() => window.location.reload()} />
  ),
  context: {
    queryClient,
  },
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

const rootElement = document.getElementById("app");

if (!rootElement) {
  throw new Error("Root element not found");
}

if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <QueryClientProvider client={queryClient}>
      <WhatspireProvider
        config={{
          baseURL: import.meta.env.VITE_SERVER_URL || "http://localhost:8080",
          apiKey: import.meta.env.VITE_API_KEY,
        }}
      >
        <RouterProvider router={router} />
      </WhatspireProvider>
    </QueryClientProvider>,
  );
}
