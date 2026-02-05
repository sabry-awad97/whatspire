/**
 * WhatsApp JID (Jabber ID) Utilities
 *
 * WhatsApp JID format: phone@s.whatsapp.net or phone:device@s.whatsapp.net
 * Examples:
 * - 201021347532@s.whatsapp.net (single device)
 * - 201021347532:98@s.whatsapp.net (multi-device)
 */

/**
 * Format JID by removing device ID but keeping domain
 * @param jid - WhatsApp JID (e.g., "201021347532:98@s.whatsapp.net")
 * @returns Formatted JID without device ID (e.g., "201021347532@s.whatsapp.net")
 */
export function formatJID(jid: string): string {
  if (!jid) return "";

  const parts = jid.split("@");
  if (parts.length !== 2) return jid;

  const userPart = parts[0].split(":")[0]; // Remove device ID
  const domain = parts[1];

  return `${userPart}@${domain}`;
}

/**
 * Extract phone number from WhatsApp JID
 * @param jid - WhatsApp JID (e.g., "201021347532:98@s.whatsapp.net")
 * @returns Phone number without device ID (e.g., "201021347532")
 */
export function extractPhoneFromJID(jid: string): string {
  if (!jid) return "";

  // Split by @ to get the user part
  const userPart = jid.split("@")[0];

  // Split by : to remove device ID if present
  const phoneNumber = userPart.split(":")[0];

  return phoneNumber;
}

/**
 * Format phone number with international prefix
 * @param jid - WhatsApp JID or phone number
 * @returns Formatted phone number with + prefix (e.g., "+201021347532")
 */
export function formatPhoneNumber(jid: string): string {
  const phone = extractPhoneFromJID(jid);
  return phone ? `+${phone}` : "";
}

/**
 * Format phone number in a more readable way
 * @param jid - WhatsApp JID or phone number
 * @returns Formatted phone number (e.g., "+20 102 134 7532")
 */
export function formatPhoneNumberReadable(jid: string): string {
  const phone = extractPhoneFromJID(jid);
  if (!phone) return "";

  // Add + prefix
  let formatted = `+${phone}`;

  // Try to format based on common patterns
  // Egypt: +20 XXX XXX XXXX
  if (phone.startsWith("20") && phone.length === 12) {
    formatted = `+20 ${phone.slice(2, 5)} ${phone.slice(5, 8)} ${phone.slice(8)}`;
  }
  // US/Canada: +1 XXX XXX XXXX
  else if (phone.startsWith("1") && phone.length === 11) {
    formatted = `+1 ${phone.slice(1, 4)} ${phone.slice(4, 7)} ${phone.slice(7)}`;
  }
  // UK: +44 XXXX XXX XXX
  else if (phone.startsWith("44") && phone.length >= 12) {
    formatted = `+44 ${phone.slice(2, 6)} ${phone.slice(6, 9)} ${phone.slice(9)}`;
  }
  // Generic: Add spaces every 3-4 digits after country code
  else if (phone.length > 10) {
    const countryCode = phone.slice(0, phone.length - 10);
    const rest = phone.slice(phone.length - 10);
    formatted = `+${countryCode} ${rest.slice(0, 3)} ${rest.slice(3, 6)} ${rest.slice(6)}`;
  }

  return formatted;
}

/**
 * Extract device ID from WhatsApp JID
 * @param jid - WhatsApp JID (e.g., "201021347532:98@s.whatsapp.net")
 * @returns Device ID or null if not present
 */
export function extractDeviceID(jid: string): string | null {
  if (!jid) return null;

  const userPart = jid.split("@")[0];
  const parts = userPart.split(":");

  return parts.length > 1 ? parts[1] : null;
}

/**
 * Check if JID is a group
 * @param jid - WhatsApp JID
 * @returns true if JID is a group
 */
export function isGroupJID(jid: string): boolean {
  return jid?.includes("@g.us") || false;
}

/**
 * Check if JID is a broadcast list
 * @param jid - WhatsApp JID
 * @returns true if JID is a broadcast list
 */
export function isBroadcastJID(jid: string): boolean {
  return jid?.includes("@broadcast") || false;
}

/**
 * Get JID type
 * @param jid - WhatsApp JID
 * @returns Type of JID (user, group, broadcast, or unknown)
 */
export function getJIDType(
  jid: string,
): "user" | "group" | "broadcast" | "unknown" {
  if (!jid) return "unknown";

  if (isGroupJID(jid)) return "group";
  if (isBroadcastJID(jid)) return "broadcast";
  if (jid.includes("@s.whatsapp.net")) return "user";

  return "unknown";
}

/**
 * Validate WhatsApp JID format
 * @param jid - WhatsApp JID to validate
 * @returns true if JID is valid
 */
export function isValidJID(jid: string): boolean {
  if (!jid) return false;

  // Check for @ symbol
  if (!jid.includes("@")) return false;

  // Check for valid domain
  const validDomains = ["s.whatsapp.net", "g.us", "broadcast"];
  const domain = jid.split("@")[1];

  return validDomains.some((d) => domain?.includes(d));
}
