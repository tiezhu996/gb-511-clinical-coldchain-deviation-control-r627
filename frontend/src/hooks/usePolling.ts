import { useCallback, useEffect, useRef } from 'react';

export function usePolling(callback: () => void | Promise<void>, intervalMs = 20000) {
  const callbackRef = useRef(callback);
  callbackRef.current = callback;
  useEffect(() => {
    const timer = window.setInterval(() => void callbackRef.current(), intervalMs);
    return () => window.clearInterval(timer);
  }, [intervalMs]);
  return useCallback(() => void callbackRef.current(), []);
}
