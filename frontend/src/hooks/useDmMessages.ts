import { useEffect, useState, useCallback } from 'react';
import pb from '@/lib/pocketbase';
import { useReconnect } from './useReconnect';

export interface DmMessage {
  id: string;
  dm: string;
  author: string;
  author_name: string;
  body: string;
  created: string;
}

const DM_PAGE_SIZE = 200;

/**
 * Messages hook for DM conversations — permanent messages.
 * Subscribes to `dm_messages` collection, filters by DM ID.
 */
export function useDmMessages(dmId: string) {
  const [messages, setMessages] = useState<DmMessage[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchMessages = useCallback(async () => {
    if (!dmId) return;
    try {
      const result = await pb
        .collection('dm_messages')
        .getList<DmMessage>(1, DM_PAGE_SIZE, {
          filter: `dm = "${dmId}"`,
          sort: 'created',
          requestKey: null,
        });
      setMessages(result.items);
    } catch (err) {
      console.error('Failed to fetch DM messages:', err);
    } finally {
      setIsLoading(false);
    }
  }, [dmId]);

  useEffect(() => {
    setIsLoading(true);
    fetchMessages();
  }, [fetchMessages]);

  // Real-time subscription
  useEffect(() => {
    if (!dmId) return;

    const unsubPromise = pb
      .collection('dm_messages')
      .subscribe<DmMessage>('*', (data) => {
        if (data.record.dm !== dmId) return;

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
  }, [dmId]);

  useReconnect(fetchMessages);

  // Optimistic send
  const sendMessage = useCallback(
    async (text: string) => {
      const tempId = `temp-${Date.now()}`;
      const optimistic: DmMessage = {
        id: tempId,
        dm: dmId,
        author: pb.authStore.record?.id ?? '',
        author_name: pb.authStore.record?.['display_name'] ?? 'Wanderer',
        body: text,
        created: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, optimistic]);

      try {
        const real = await pb.collection('dm_messages').create<DmMessage>({
          dm: dmId,
          author: pb.authStore.record?.id,
          body: text,
        });
        setMessages((prev) =>
          prev.map((m) => (m.id === tempId ? real : m))
        );
      } catch {
        setMessages((prev) => prev.filter((m) => m.id !== tempId));
      }
    },
    [dmId]
  );

  return { messages, isLoading, sendMessage };
}
