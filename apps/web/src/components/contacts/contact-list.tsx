import { useState } from "react";
import { RefreshCw, Search, UserPlus } from "lucide-react";

import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { ContactCard } from "./contact-card";

// ============================================================================
// Types
// ============================================================================

export interface Contact {
  id: string;
  name: string;
  phoneNumber: string;
  profilePicture?: string;
  status?: string;
  lastSeen?: string;
  isBlocked?: boolean;
}

interface ContactListProps {
  sessionId?: string;
  onSync?: () => void;
}

// ============================================================================
// Mock Data
// ============================================================================

const MOCK_CONTACTS: Contact[] = [
  {
    id: "1",
    name: "John Doe",
    phoneNumber: "+1234567890",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=John",
    status: "Hey there! I am using WhatsApp",
    lastSeen: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
  },
  {
    id: "2",
    name: "Jane Smith",
    phoneNumber: "+9876543210",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=Jane",
    status: "Busy",
    lastSeen: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: "3",
    name: "Bob Johnson",
    phoneNumber: "+1122334455",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=Bob",
    status: "Available",
    lastSeen: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
  },
  {
    id: "4",
    name: "Alice Williams",
    phoneNumber: "+5544332211",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=Alice",
    lastSeen: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
  },
  {
    id: "5",
    name: "Charlie Brown",
    phoneNumber: "+6677889900",
    profilePicture: "https://api.dicebear.com/7.x/avataaars/svg?seed=Charlie",
    status: "At work",
    isBlocked: true,
  },
];

// ============================================================================
// Component
// ============================================================================

export function ContactList({ sessionId, onSync }: ContactListProps) {
  const [contacts, setContacts] = useState<Contact[]>(MOCK_CONTACTS);
  const [searchQuery, setSearchQuery] = useState("");
  const [isSyncing, setIsSyncing] = useState(false);

  const handleSync = async () => {
    setIsSyncing(true);
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1500));
    setIsSyncing(false);
    onSync?.();
  };

  // Filter contacts based on search
  const filteredContacts = contacts.filter(
    (contact) =>
      contact.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      contact.phoneNumber.includes(searchQuery),
  );

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="glass-card-enhanced p-4 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold">Contacts</h2>
            <p className="text-sm text-muted-foreground">
              {filteredContacts.length} of {contacts.length} contacts
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleSync}
              disabled={isSyncing}
              className="glass-card hover-glow-teal"
            >
              <RefreshCw
                className={`h-4 w-4 mr-2 ${isSyncing ? "animate-spin" : ""}`}
              />
              {isSyncing ? "Syncing..." : "Sync Contacts"}
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="glass-card hover-glow-emerald"
            >
              <UserPlus className="h-4 w-4 mr-2" />
              Add Contact
            </Button>
          </div>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search contacts by name or phone..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="glass-card pl-10"
          />
        </div>
      </div>

      {/* Contact List */}
      <div className="flex-1 overflow-y-auto p-4">
        {filteredContacts.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <svg
              className="h-16 w-16 text-muted-foreground mb-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
            <h3 className="text-lg font-semibold mb-2">No contacts found</h3>
            <p className="text-sm text-muted-foreground mb-4">
              {searchQuery
                ? "Try adjusting your search query"
                : "Sync your contacts to get started"}
            </p>
            {!searchQuery && (
              <Button
                onClick={handleSync}
                disabled={isSyncing}
                className="glass-card hover-glow-teal"
              >
                <RefreshCw
                  className={`h-4 w-4 mr-2 ${isSyncing ? "animate-spin" : ""}`}
                />
                Sync Contacts
              </Button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredContacts.map((contact) => (
              <ContactCard key={contact.id} contact={contact} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
