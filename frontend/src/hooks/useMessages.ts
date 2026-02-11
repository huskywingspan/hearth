import { useEffect, useState, useCallback } from 'react';
import pb from '@/lib/pocketbase';
import { useReconnect } from './useReconnect';
import { MESSAGE_PAGE_SIZE } from '@/lib/constants';

export interface Message {
  id: string;
  body: string;
  room: string;
  author: string;
  author_name: string;
  expires_at: string;
  created: string;
  expand?: {
    author?: { display_name: string; avatar_url: string };
  };
}

export function useMessages(roomId: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch current messages for this room
  const fetchMessages = useCallback(async () => {
    if (!roomId) return;
    try {
      const result = await pb
        .collection('messages')
        .getList<Message>(1, MESSAGE_PAGE_SIZE, {
          filter: `room = "${roomId}"`,
          sort: 'created',
          expand: 'author',
          requestKey: `messages-${roomId}`,
        });
      setMessages(result.items);
    } catch (err) {
      console.error('Failed to fetch messages:', err);
    } finally {
      setIsLoading(false);
    }
  }, [roomId]);

  // Initial fetch
  useEffect(() => {
    setIsLoading(true);
    fetchMessages();
  }, [fetchMessages]);

  // Real-time subscription
  useEffect(() => {
    if (!roomId) return;

    const unsubPromise = pb
      .collection('messages')
      .subscribe<Message>('*', (data) => {
        if (data.record.room !== roomId) return;

        switch (data.action) {
          case 'create':
            setMessages((prev) => {
              // Skip own messages — the optimistic send flow handles them
              if (data.record.author === pb.authStore.record?.id) return prev;
              // Also deduplicate by id just in case
              if (prev.some((m) => m.id === data.record.id)) return prev;
              return [...prev, data.record];
            });
            break;
          case 'update':
            setMessages((prev) =>
              prev.map((m) => (m.id === data.record.id ? data.record : m))
            );
            break;
          case 'delete':
            // Server GC deleted an expired message — remove from state
            setMessages((prev) =>
              prev.filter((m) => m.id !== data.record.id)
            );
            break;
        }
      });

    return () => {
      unsubPromise.then((unsub) => unsub());
    };
  }, [roomId]);

  // Reconnect: re-fetch everything (missed events are not replayed per R-004)
  useReconnect(fetchMessages);

  // Remove a message client-side (e.g., after animationend)
  const removeMessage = useCallback((id: string) => {
    setMessages((prev) => prev.filter((m) => m.id !== id));
  }, []);

  // Optimistic send
  const sendMessage = useCallback(
    async (text: string) => {
      const tempId = `temp-${Date.now()}`;
      const optimistic: Message = {
        id: tempId,
        body: text,
        room: roomId,
        author: pb.authStore.record?.id ?? '',
        author_name: pb.authStore.record?.['display_name'] ?? 'Wanderer',
        expires_at: new Date(Date.now() + 300_000).toISOString(), // 5 min placeholder
        created: new Date().toISOString(),
      };

      // Immediately add to state (optimistic UI)
      setMessages((prev) => [...prev, optimistic]);

      try {
        const real = await pb.collection('messages').create<Message>({
          body: text,
          room: roomId,
          author: pb.authStore.record?.id,
          // Server enforces expires_at via TTL hook
        });
        // Replace optimistic with real record
        setMessages((prev) =>
          prev.map((m) => (m.id === tempId ? real : m))
        );
      } catch {
        // Revert on failure
        setMessages((prev) => prev.filter((m) => m.id !== tempId));
        // TODO: show error toast via toast context
      }
    },
    [roomId]
  );

  return { messages, isLoading, sendMessage, removeMessage };
}
