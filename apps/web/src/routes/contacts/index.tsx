import { createFileRoute } from "@tanstack/react-router";
import { toast } from "sonner";

import { ContactList } from "@/components/contacts/contact-list";

export const Route = createFileRoute("/contacts/")({
  component: ContactsComponent,
});

function ContactsComponent() {
  const handleSync = () => {
    toast.success("Contacts synced successfully");
  };

  return (
    <div className="h-screen network-bg flex flex-col">
      <div className="flex-1 max-w-7xl mx-auto w-full p-6 flex flex-col">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold gradient-text">Contacts</h1>
          <p className="text-muted-foreground">Manage your WhatsApp contacts</p>
        </div>

        {/* Contact List */}
        <div className="flex-1 glass-card-enhanced rounded-lg overflow-hidden">
          <ContactList onSync={handleSync} />
        </div>
      </div>
    </div>
  );
}
