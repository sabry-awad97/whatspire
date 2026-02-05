import { createFileRoute } from "@tanstack/react-router";

import { GroupDetails } from "@/components/groups/group-details";

export const Route = createFileRoute("/groups/$groupId")({
  component: GroupDetailsPage,
});

function GroupDetailsPage() {
  const { groupId } = Route.useParams();

  // Mock group data - in real app, fetch from API
  const mockGroup = {
    id: groupId,
    name: "Family Group",
    description: "Our lovely family",
    profilePicture: "https://api.dicebear.com/7.x/identicon/svg?seed=family",
    participantCount: 8,
    isAdmin: true,
  };

  return <GroupDetails group={mockGroup} />;
}
