import { useEffect, useState } from 'react';
import pb from '@/lib/pocketbase';
import { HEARTBEAT_INTERVAL_MS, PRESENCE_POLL_INTERVAL_MS } from '@/lib/constants';

export interface PresenceEntry {
  user_id: string;
  display_name: string;
  last_seen: number;
}

export function usePresence(roomId: string) {
  const [presentUsers, setPresentUsers] = useState<PresenceEntry[]>([]);

  // Send heartbeat every 30 seconds
  useEffect(() => {
    if (!roomId) return;

    const beat = () => {
      pb.send('/api/hearth/presence/heartbeat', {
        method: 'POST',
        body: { room_id: roomId },
      }).catch(() => {
        /* offline — will retry on reconnect */
      });
    };

    beat(); // Immediate first heartbeat
    const interval = setInterval(beat, HEARTBEAT_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [roomId]);

  // Poll presence list
  useEffect(() => {
    if (!roomId) return;

    const fetchPresence = async () => {
      try {
        const data = await pb.send(`/api/hearth/presence/${roomId}`, {
          method: 'GET',
        });
        setPresentUsers(
          (data as { online?: PresenceEntry[] }).online ?? []
        );
      } catch {
        /* noop */
      }
    };

    fetchPresence();
    const interval = setInterval(fetchPresence, PRESENCE_POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [roomId]);

  return { presentUsers };
}
