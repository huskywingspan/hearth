import { useParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { Shell } from '@/components/layout/Shell';
import { DmRoom } from '@/components/dm/DmRoom';
import { Spinner } from '@/components/ui/Spinner';
import pb from '@/lib/pocketbase';

interface DmRecord {
  id: string;
  participant_a: string;
  participant_b: string;
  expand?: {
    participant_a?: { display_name: string };
    participant_b?: { display_name: string };
  };
}

export default function DmPage() {
  const { dmId } = useParams<{ dmId: string }>();
  const [dm, setDm] = useState<DmRecord | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!dmId) return;

    pb.collection('direct_messages')
      .getOne<DmRecord>(dmId, { expand: 'participant_a,participant_b', requestKey: null })
      .then(setDm)
      .catch(() => setError('Conversation not found'));
  }, [dmId]);

  if (error) {
    return (
      <Shell>
        <div className="flex items-center justify-center h-full text-[var(--color-alert-clay)]">
          {error}
        </div>
      </Shell>
    );
  }

  if (!dm || !dmId) {
    return (
      <Shell>
        <div className="flex items-center justify-center h-full">
          <Spinner />
        </div>
      </Shell>
    );
  }

  // Determine the other participant's name
  const myId = pb.authStore.record?.id;
  const otherUserName =
    dm.participant_a === myId
      ? dm.expand?.participant_b?.display_name || 'User'
      : dm.expand?.participant_a?.display_name || 'User';

  return (
    <Shell>
      <DmRoom dmId={dmId} otherUserName={otherUserName} />
    </Shell>
  );
}
