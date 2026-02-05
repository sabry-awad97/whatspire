import { createFileRoute } from "@tanstack/react-router";

import { MessageList } from "@/components/messages/message-list";

export const Route = createFileRoute("/messages/")({
  component: MessagesComponent,
});

function MessagesComponent() {
  return (
    <div className="h-screen network-bg flex flex-col">
      <div className="flex-1 max-w-7xl mx-auto w-full p-6 flex flex-col">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold gradient-text">Messages</h1>
          <p className="text-muted-foreground">
            View and manage your WhatsApp messages
          </p>
        </div>

        {/* Message List */}
        <div className="flex-1 glass-card-enhanced rounded-lg overflow-hidden">
          <MessageList />
        </div>
      </div>
    </div>
  );
}
