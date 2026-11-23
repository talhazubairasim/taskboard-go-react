import { useState, type FormEvent } from 'react';
import { useMutation } from '@apollo/client';
import { UPDATE_PROFILE } from '../../graphql/mutations';
import { useAuth } from '../../contexts/AuthContext';
import type { User } from '../../types';
import { Avatar } from '../common/Avatar';
import { getErrorMessage } from '../../utils/errors';

const inputClass =
  'block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500';

/** Lets the signed-in user change their name and avatar. */
export const UserProfile = () => {
  const { user, updateUser } = useAuth();
  const [name, setName] = useState(user?.name ?? '');
  const [avatar, setAvatar] = useState(user?.avatar ?? '');
  const [message, setMessage] = useState<{ ok: boolean; text: string } | null>(null);
  const [mutate, { loading }] = useMutation<{ updateProfile: User }>(UPDATE_PROFILE);

  if (!user) return null;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setMessage({ ok: false, text: 'Name is required' });
      return;
    }
    try {
      const { data } = await mutate({ variables: { name: name.trim(), avatar: avatar.trim() } });
      if (data) updateUser({ ...user, ...data.updateProfile });
      setMessage({ ok: true, text: 'Profile saved' });
    } catch (err) {
      setMessage({ ok: false, text: getErrorMessage(err) });
    }
  };

  return (
    <div className="max-w-lg rounded-lg border border-gray-200 bg-white p-6">
      <div className="mb-6 flex items-center gap-4">
        <Avatar name={name || user.name} src={avatar.trim() || null} size="lg" />
        <div>
          <p className="text-lg font-semibold text-gray-900">{user.name}</p>
          <p className="text-sm text-gray-500">{user.email}</p>
        </div>
      </div>
      <form onSubmit={handleSubmit} className="space-y-4" noValidate>
        {message && (
          <div role={message.ok ? 'status' : 'alert'}
            className={`rounded-md p-3 text-sm ${message.ok ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'}`}>
            {message.text}
          </div>
        )}
        <div>
          <label htmlFor="profile-name" className="mb-1 block text-sm font-medium text-gray-700">Name</label>
          <input id="profile-name" className={inputClass} value={name} onChange={(e) => setName(e.target.value)} />
        </div>
        <div>
          <label htmlFor="profile-avatar" className="mb-1 block text-sm font-medium text-gray-700">Avatar URL</label>
          <input id="profile-avatar" type="url" placeholder="https://..." className={inputClass} value={avatar}
            onChange={(e) => setAvatar(e.target.value)} />
          <p className="mt-1 text-xs text-gray-500">Leave empty to use your initial.</p>
        </div>
        <button type="submit" disabled={loading}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50">
          {loading ? 'Saving...' : 'Save profile'}
        </button>
      </form>
    </div>
  );
};
