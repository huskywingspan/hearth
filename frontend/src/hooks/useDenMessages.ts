import { useEffect, useState, useCallback } from 'react';
import pb from '@/lib/pocketbase';
import { useReconnect } from './useReconnect';
import type { Message } from './useMessages';

/** Page size for Den message history (larger — permanent messages accumulate). */
const DEN_PAGE_SIZE = 200;

/**
 * Messages hook for Dens — permanent messages, no fade.
 * Similar to useMessages but:
 * - No TTL/expiry logic (messages persist forever)
 * - No animationend cleanup
 * - Same optimistic send pattern
 */
export function useDenMessages(roomId: string) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchMessages = useCallback(async () => {
    if (!roomId) return;
    try {
      const result = await pb
        .collection('messages')
        .getList<Message>(1, DEN_PAGE_SIZE, {
          filter: `room = "${roomId}"`,
          sort: 'created',
          expand: 'author',
          requestKey: null,
        });
      setMessages(result.items);
    } catch (err) {
      console.error('Failed to fetch den messages:', err);
    } finally {
      setIsLoading(false);
    }
  }, [roomId]);

  useEffect(() => {
    setIsLoading(true);
    fetchMessages();
  }, [fetchMessages]);

  // Real-time subscription (same collection as campfire messages)
  useEffect(() => {
    if (!roomId) return;

    const unsubPromise = pb
      .collection('messages')
      .subscribe<Message>('*', (data) => {
        if (data.record.room !== roomId) return;

        switch (data.action) {
          case 'create':
            setMessages((prev) => {
              if (data.record.author === pb.authStore.record?.id) return prev;
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

  useReconnect(fetchMessages);

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
        expires_at: '2099-12-31T23:59:59Z', // Den messages don't expire
        created: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, optimistic]);

      try {
        const real = await pb.collection('messages').create<Message>({
          body: text,
          room: roomId,
          author: pb.authStore.record?.id,
        });
        setMessages((prev) =>
          prev.map((m) => (m.id === tempId ? real : m))
        );
      } catch {
        setMessages((prev) => prev.filter((m) => m.id !== tempId));
      }
    },
    [roomId]
  );

  return { messages, isLoading, sendMessage };
}
