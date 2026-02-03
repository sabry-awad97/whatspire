import { AlertTriangle, Home, RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

interface ErrorFallbackProps {
  /** The error that was caught */
  error?: Error | unknown;
  /** Callback to reset/retry */
  onReset?: () => void;
  /** Title to display */
  title?: string;
  /** Description message */
  description?: string;
}

/**
 * User-friendly error fallback component for error boundaries.
 *
 * Displays:
 * - Error icon and message
 * - Retry button to attempt recovery
 * - Home link for navigation fallback
 * - Stack trace in development mode
 *
 * @example
 * ```tsx
 * <ErrorBoundary fallback={<ErrorFallback />}>
 *   <MyComponent />
 * </ErrorBoundary>
 * ```
 */
export function ErrorFallback({
  error,
  onReset,
  title = "Something went wrong",
  description = "An unexpected error occurred. Please try again or return to the home page.",
}: ErrorFallbackProps) {
  const errorMessage =
    error instanceof Error ? error.message : "Unknown error occurred";

  const errorStack =
    import.meta.env.DEV && error instanceof Error ? error.stack : null;

  const handleReset = () => {
    if (onReset) {
      onReset();
    } else {
      // Fallback: reload the current page
      window.location.reload();
    }
  };

  const handleGoHome = () => {
    window.location.href = "/";
  };

  return (
    <div className="flex min-h-[400px] w-full items-center justify-center p-4">
      <Card className="w-full max-w-md p-6">
        <div className="flex flex-col items-center text-center">
          {/* Error Icon */}
          <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
            <AlertTriangle className="h-6 w-6 text-destructive" />
          </div>

          {/* Title */}
          <h2 className="mb-2 text-xl font-semibold">{title}</h2>

          {/* Description */}
          <p className="mb-4 text-sm text-muted-foreground">{description}</p>

          {/* Error Message */}
          <div className="mb-6 w-full rounded border border-destructive/20 bg-destructive/5 p-3">
            <p className="text-sm font-mono text-destructive">{errorMessage}</p>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3">
            <Button variant="outline" onClick={handleGoHome}>
              <Home className="mr-2 h-4 w-4" />
              Go Home
            </Button>
            <Button onClick={handleReset}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Try Again
            </Button>
          </div>

          {/* Stack Trace (Development Only) */}
          {errorStack && (
            <details className="mt-6 w-full text-left">
              <summary className="cursor-pointer text-xs text-muted-foreground hover:text-foreground">
                Show stack trace (dev only)
              </summary>
              <pre className="mt-2 max-h-48 overflow-auto rounded bg-muted p-3 text-xs">
                {errorStack}
              </pre>
            </details>
          )}
        </div>
      </Card>
    </div>
  );
}

/**
 * Minimal inline error fallback for smaller components
 */
export function ErrorFallbackInline({
  error,
  onReset,
}: Pick<ErrorFallbackProps, "error" | "onReset">) {
  const errorMessage =
    error instanceof Error ? error.message : "Something went wrong";

  return (
    <div className="flex items-center gap-2 rounded border border-destructive/20 bg-destructive/5 p-2 text-sm">
      <AlertTriangle className="h-4 w-4 text-destructive" />
      <span className="flex-1 text-destructive">{errorMessage}</span>
      {onReset && (
        <Button variant="ghost" size="xs" onClick={onReset}>
          <RefreshCw className="h-3 w-3" />
        </Button>
      )}
    </div>
  );
}
