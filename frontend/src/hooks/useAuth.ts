
import { useCallback, useEffect, useState } from 'react';
import { AUTH_EVENT, clearSession, getSession, saveSession } from '../api/client';
import { currentSession, login } from '../api/auth';
import type { UserSession } from '../types/domain';

export function useAuth() {
	const [session, setSession] = useState<UserSession | null>(null);
	const [loading, setLoading] = useState(true);
	useEffect(() => {
		const expired = () => setSession(null);
		window.addEventListener(AUTH_EVENT, expired);
		const cached = getSession();
		if (!cached) setLoading(false);
		else currentSession().then((profile) => { const next = { ...cached, ...profile }; saveSession(next); setSession(next); }).catch(() => clearSession()).finally(() => setLoading(false));
		return () => window.removeEventListener(AUTH_EVENT, expired);
	}, []);
  const signIn = useCallback(async (username: string, password: string) => {
	setLoading(true);
	try { const next = await login(username, password); saveSession(next); setSession(next); return next; }
	finally { setLoading(false); }
  }, []);
  const logout = () => { clearSession(); setSession(null); };
  return { session, loading, authenticated: Boolean(session), signIn, logout };
}
