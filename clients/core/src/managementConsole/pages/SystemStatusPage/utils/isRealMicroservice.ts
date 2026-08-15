/**
 * Course phase types that are handled directly by the core server (e.g. Application, Matching,
 * DevOps Challenge) have `baseUrl` set to "core" instead of a real service URL
 * This function decides one from the other.
 */
export function isRealMicroservice(baseUrl: string): boolean {
  try {
    const parsed = new URL(baseUrl);
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
}
