import {useState, useRef, useMemo, useEffect, useCallback} from 'react';
import type {SaveState} from '@/components/common/EditPageHeader';

// ============================================================================
// useDirtyState
// ============================================================================

/**
 * Dirty state tracking hook return type
 */
export interface UseDirtyStateReturn<T> {
  /** Current form data */
  form: T;
  /** Update form data */
  setForm: React.Dispatch<React.SetStateAction<T>>;
  /** Whether there are unsaved changes */
  isDirty: boolean;
  /** Reset dirty state (call after successful save) */
  resetDirty: () => void;
  /** Sync form from server data and reset dirty state */
  syncFromData: (data: T) => void;
}

/**
 * useDirtyState - Tracks whether form data has changed from initial values
 * Uses JSON.stringify deep comparison for dirty detection
 * Supports beforeunload interception to prevent accidental navigation
 *
 * @param initialData - Initial form data
 * @param options - Configuration options
 */
export function useDirtyState<T extends Record<string, any>>(
  initialData: T,
  options?: { interceptBeforeUnload?: boolean }
): UseDirtyStateReturn<T> {
  const {interceptBeforeUnload = true} = options || {};
  const [form, setForm] = useState<T>(initialData);
  const initialFormRef = useRef<T>(initialData);

  const isDirty = useMemo(() => {
    return JSON.stringify(form) !== JSON.stringify(initialFormRef.current);
  }, [form]);

  const resetDirty = useCallback(() => {
    initialFormRef.current = {...form};
  }, [form]);

  const syncFromData = useCallback((data: T) => {
    setForm(data);
    initialFormRef.current = {...data};
  }, []);

  // beforeunload interception
  useEffect(() => {
    if (!interceptBeforeUnload) return;

    const handler = (e: BeforeUnloadEvent) => {
      if (isDirty) {
        e.preventDefault();
        e.returnValue = '';
      }
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [isDirty, interceptBeforeUnload]);

  return {form, setForm, isDirty, resetDirty, syncFromData};
}

// ============================================================================
// useSaveState
// ============================================================================

/**
 * useSaveState - Manages save operation state with auto-reset timers
 * idle -> saving -> success (2s) -> idle
 * idle -> saving -> error (3s) -> idle
 */
export function useSaveState() {
  const [saveState, setSaveState] = useState<SaveState>('idle');
  const timerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const setSaving = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    setSaveState('saving');
  }, []);

  const setSuccess = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    setSaveState('success');
    timerRef.current = setTimeout(() => setSaveState('idle'), 2000);
  }, []);

  const setError = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    setSaveState('error');
    timerRef.current = setTimeout(() => setSaveState('idle'), 3000);
  }, []);

  const reset = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    setSaveState('idle');
  }, []);

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  return {
    saveState,
    isSaving: saveState === 'saving',
    setSaving,
    setSuccess,
    setError,
    reset,
  };
}

// ============================================================================
// useKeyboardShortcut
// ============================================================================

/**
 * useKeyboardShortcut - Registers a keyboard shortcut listener
 *
 * @param key - Key combination (e.g., 'ctrl+s', 'meta+s')
 * @param callback - Function to call when shortcut is triggered
 * @param options - Configuration options
 */
export function useKeyboardShortcut(
  key: string,
  callback: () => void,
  options?: {
    preventDefault?: boolean;
    enabled?: boolean;
  }
) {
  const {preventDefault = true, enabled = true} = options || {};

  useEffect(() => {
    if (!enabled) return;

    const parts = key.toLowerCase().split('+');
    const targetKey = parts[parts.length - 1];
    const needsCtrl = parts.includes('ctrl') || parts.includes('control');
    const needsMeta = parts.includes('meta') || parts.includes('cmd') || parts.includes('command');
    const needsShift = parts.includes('shift');
    const needsAlt = parts.includes('alt');

    const handler = (e: KeyboardEvent) => {
      const ctrlMatch = needsCtrl ? (e.ctrlKey || e.metaKey) : !e.ctrlKey && !e.metaKey;
      const metaMatch = needsMeta ? e.metaKey : true;
      const shiftMatch = needsShift ? e.shiftKey : !e.shiftKey;
      const altMatch = needsAlt ? e.altKey : !e.altKey;

      if (
        e.key.toLowerCase() === targetKey &&
        ctrlMatch &&
        metaMatch &&
        shiftMatch &&
        altMatch
      ) {
        if (preventDefault) e.preventDefault();
        callback();
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [key, callback, preventDefault, enabled]);
}
