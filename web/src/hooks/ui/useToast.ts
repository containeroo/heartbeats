import { useCallback, useRef, useState } from "react";

/** useToast manages toast messages and auto-dismiss timers. */
export function useToast() {
  const [toast, setToast] = useState<string | null>(null);
  const timerRef = useRef<number | null>(null);

  const show = useCallback((message: string, ms = 2400) => {
    setToast(message);
    if (timerRef.current) {
      window.clearTimeout(timerRef.current);
    }
    timerRef.current = window.setTimeout(() => {
      setToast(null);
      timerRef.current = null;
    }, ms);
  }, []);

  const clear = useCallback(() => {
    if (timerRef.current) {
      window.clearTimeout(timerRef.current);
      timerRef.current = null;
    }
    setToast(null);
  }, []);

  return { toast, show, clear };
}
