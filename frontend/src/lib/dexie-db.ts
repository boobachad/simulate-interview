import Dexie, { type EntityTable } from "dexie";
import type { CombinedStats, UserID } from "@/lib/api";

// Stats cache entry
interface StatsCache {
  id?: number;
  user_id: UserID;
  platform: "combined";
  data: CombinedStats;
  cached_at: number; // Unix timestamp in milliseconds
}

// Dexie database class
class InterviewDatabase extends Dexie {
  stats!: EntityTable<StatsCache, "id">;

  constructor() {
    super("InterviewDatabase");

    this.version(1).stores({
      stats: "++id, user_id, platform, cached_at",
    });
  }
}

// Create database instance
export const db = new InterviewDatabase();

// Cache TTL: 1 hour in milliseconds
const CACHE_TTL_MS = 60 * 60 * 1000;

// Check if cached data is still valid
function isCacheValid(cachedAt: number): boolean {
  const now = Date.now();
  return now - cachedAt < CACHE_TTL_MS;
}

// Stats cache operations
export const statsCache = {
  // Get cached stats for user
  async get(userId: UserID): Promise<CombinedStats | null> {
    const cached = await db.stats
      .where("user_id")
      .equals(userId)
      .and((entry) => entry.platform === "combined")
      .first();

    if (!cached) return null;

    // Check if cache is still valid
    if (!isCacheValid(cached.cached_at)) {
      // Delete expired cache
      await db.stats.delete(cached.id!);
      return null;
    }

    return cached.data;
  },

  // Set cached stats for user
  async set(userId: UserID, data: CombinedStats): Promise<void> {
    // Use transaction to prevent race conditions
    await db.transaction("rw", db.stats, async () => {
      // Delete existing cache for this user
      await db.stats
        .where("user_id")
        .equals(userId)
        .and((entry) => entry.platform === "combined")
        .delete();

      // Insert new cache entry
      await db.stats.add({
        user_id: userId,
        platform: "combined",
        data,
        cached_at: Date.now(),
      });
    });
  },

  // Clear all cached stats
  async clear(): Promise<void> {
    await db.stats.clear();
  },

  // Clear expired cache entries
  async clearExpired(): Promise<void> {
    const now = Date.now();
    await db.stats.where("cached_at").below(now - CACHE_TTL_MS).delete();
  },
};

// Initialize: clear expired cache on load
if (typeof window !== "undefined") {
  statsCache.clearExpired().catch((err) => {
    console.error("Failed to clear expired cache:", err);
  });
}
