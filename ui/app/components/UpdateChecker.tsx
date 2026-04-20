"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { apiClient } from "@/lib/apiClient";
import { useWS } from "@/components/WebSocketProvider";

const DISMISS_KEY = "update_dismissed_version";
const CHECK_INTERVAL_MS = 30 * 60 * 1000; // 30분
const POLL_INTERVAL_MS = 2000; // WS 없을 때 폴링 간격

type UpdatePhase =
  | 'idle'
  | 'checking'
  | 'downloading'
  | 'downloaded'
  | 'applying'
  | 'restart_required'
  | 'error';

interface UpdateState {
  hasUpdate: boolean;
  current: string;
  latest: string;
  updateUrl: string;
  notes: string;
  phase: UpdatePhase;
  percent: number;
  totalBytes: number;
  downloadedBytes: number;
  error: string;
}

const INITIAL_STATE: UpdateState = {
  hasUpdate: false,
  current: '',
  latest: '',
  updateUrl: '',
  notes: '',
  phase: 'idle',
  percent: 0,
  totalBytes: 0,
  downloadedBytes: 0,
  error: '',
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

export default function UpdateChecker() {
  const [state, setState] = useState<UpdateState>(INITIAL_STATE);
  const [dismissed, setDismissed] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const { subscribe, isOpen } = useWS();

  const checkUpdate = useCallback(async () => {
    try {
      const data = await apiClient.checkUpdate();
      if (!data) return;
      setState(prev => ({
        ...prev,
        hasUpdate: data.hasUpdate ?? false,
        current: data.current ?? prev.current,
        latest: data.latest ?? prev.latest,
        updateUrl: data.updateUrl ?? prev.updateUrl,
        notes: data.notes ?? prev.notes,
      }));
      if (data.latest) {
        const saved = localStorage.getItem(DISMISS_KEY);
        if (saved === data.latest) setDismissed(true);
      }
    } catch {
      // 네트워크 오류 등은 조용히 무시
    }
  }, []);

  useEffect(() => {
    checkUpdate();
    const timer = setInterval(checkUpdate, CHECK_INTERVAL_MS);
    return () => clearInterval(timer);
  }, [checkUpdate]);

  useEffect(() => {
    const unsubscribe = subscribe((msg) => {
      if (msg.type === 'update_progress') {
        setState(prev => ({
          ...prev,
          phase: (msg.state as UpdatePhase) || prev.phase,
          percent: msg.percent ?? prev.percent,
          totalBytes: msg.totalBytes ?? prev.totalBytes,
          downloadedBytes: msg.downloadedBytes ?? prev.downloadedBytes,
          latest: msg.version || prev.latest,
          error: msg.error || '',
        }));
      }
    });
    return unsubscribe;
  }, [subscribe]);

  useEffect(() => {
    const isActivePhase = (phase: UpdatePhase) =>
      phase === 'downloading' || phase === 'applying';

    if (!isOpen && isActivePhase(state.phase)) {
      pollRef.current = setInterval(async () => {
        try {
          const status = await apiClient.getUpdateStatus();
          setState(prev => ({
            ...prev,
            phase: status.state as UpdatePhase,
            percent: status.percent ?? prev.percent,
            totalBytes: status.totalBytes ?? prev.totalBytes,
            downloadedBytes: status.downloadedBytes ?? prev.downloadedBytes,
            latest: status.version || prev.latest,
            error: status.error || '',
          }));
        } catch {
          // 무시
        }
      }, POLL_INTERVAL_MS);
    } else {
      if (pollRef.current) {
        clearInterval(pollRef.current);
        pollRef.current = null;
      }
    }

    return () => {
      if (pollRef.current) {
        clearInterval(pollRef.current);
        pollRef.current = null;
      }
    };
  }, [isOpen, state.phase]);

  const handleStartDownload = async () => {
    setState(s => ({ ...s, phase: 'downloading', percent: 0, error: '' }));
    try {
      const result = await apiClient.startUpdateDownload();
      if (!result.ok) {
        setState(s => ({ ...s, phase: 'error', error: result.error || '다운로드 시작 실패' }));
      }
    } catch (e: any) {
      setState(s => ({ ...s, phase: 'error', error: e?.message || '다운로드 시작 실패' }));
    }
  };

  const handleApply = async () => {
    setState(s => ({ ...s, phase: 'applying' }));
    try {
      const result = await apiClient.applyUpdate();
      if (result.ok) {
        setState(s => ({ ...s, phase: 'restart_required' }));
      } else {
        setState(s => ({ ...s, phase: 'error', error: result.error || '적용 실패' }));
      }
    } catch (e: any) {
      setState(s => ({ ...s, phase: 'error', error: e?.message || '적용 실패' }));
    }
  };

  const handleCancel = async () => {
    await apiClient.cancelUpdateDownload();
    setState(s => ({ ...s, phase: 'idle', percent: 0 }));
  };

  const handleDismiss = () => {
    if (state.latest) localStorage.setItem(DISMISS_KEY, state.latest);
    setDismissed(true);
  };

  const handleRetry = () => {
    setState(s => ({ ...s, phase: 'idle', error: '' }));
    checkUpdate();
  };

  if (process.env.NODE_ENV === 'development') return null;
  if (state.phase === 'idle' && (!state.hasUpdate || dismissed)) return null;

  const { phase, percent, totalBytes, downloadedBytes, latest, current, updateUrl, error } = state;

  const bannerBase =
    "w-full flex items-center gap-2.5 px-5 py-2 text-[13px] box-border flex-wrap border-b";
  const bannerBg =
    phase !== 'idle'
      ? "bg-[var(--surface-elevated,#f8fafc)] border-[var(--accent-border,#c7d2fe)]"
      : "bg-[var(--accent-light,#eef2ff)] border-[var(--accent-border,#c7d2fe)]";

  const badgeBase =
    "inline-flex items-center px-2 py-0.5 rounded-[10px] text-[11px] font-bold tracking-wide flex-shrink-0 whitespace-nowrap text-white";

  const btnBase =
    "px-3 py-0.5 rounded-md text-xs font-semibold cursor-pointer flex-shrink-0 border-none transition-colors whitespace-nowrap";

  const linkBase =
    "px-3 py-0.5 rounded-md bg-transparent text-xs font-semibold flex-shrink-0 border border-[var(--accent-border,#c7d2fe)] no-underline whitespace-nowrap transition-colors hover:bg-[var(--accent-light,#eef2ff)]";

  return (
    <div className={`${bannerBase} ${bannerBg}`}>

      {/* 알림 상태 */}
      {phase === 'idle' && state.hasUpdate && (
        <>
          <span className={`${badgeBase} bg-[var(--accent,#1f3f62)]`}>업데이트 가능</span>
          <span className="flex-1 min-w-0 text-[var(--text-secondary)] whitespace-nowrap overflow-hidden text-ellipsis">
            현재 버전{' '}
            <span className="font-semibold text-[var(--text-primary)] font-mono">{current}</span>
            {' '}→{' '}
            <span className="font-semibold text-[var(--accent,#1f3f62)] font-mono">{latest}</span>
          </span>
          <button
            className={`${btnBase} bg-[var(--accent,#1f3f62)] text-white hover:bg-[var(--accent-hover,#2d5a8a)]`}
            onClick={handleStartDownload}
          >
            지금 업데이트
          </button>
          {updateUrl && (
            <a
              href={updateUrl}
              target="_blank"
              rel="noopener noreferrer"
              className={`${linkBase} text-[var(--accent,#1f3f62)]`}
            >
              릴리스 노트
            </a>
          )}
          <button
            className="bg-transparent border-none text-lg leading-none text-[var(--text-muted)] cursor-pointer px-0.5 flex-shrink-0 ml-auto hover:text-[var(--error,#dc2626)] transition-colors"
            onClick={handleDismiss}
            title="닫기"
          >
            &times;
          </button>
        </>
      )}

      {/* 다운로드 중 */}
      {phase === 'downloading' && (
        <>
          <span className={`${badgeBase} bg-[#d97706]`}>다운로드 중...</span>
          <div className="flex-1 flex items-center gap-2.5 min-w-0">
            <div className="flex-1 h-1.5 bg-[var(--surface-input,#e2e8f0)] rounded-full overflow-hidden min-w-[80px]">
              <div
                className="h-full bg-[var(--accent,#1f3f62)] rounded-full transition-[width] duration-300"
                style={{ width: `${percent}%` }}
              />
            </div>
            <span className="text-xs text-[var(--text-secondary)] whitespace-nowrap flex-shrink-0 tabular-nums">
              {totalBytes > 0
                ? `${formatBytes(downloadedBytes)} / ${formatBytes(totalBytes)} (${Math.round(percent)}%)`
                : `${Math.round(percent)}%`}
            </span>
          </div>
          <button
            className={`${btnBase} bg-transparent text-[var(--text-secondary)] border border-[var(--accent-border,#c7d2fe)] hover:bg-[var(--accent-light,#eef2ff)] hover:text-[var(--text-primary)]`}
            onClick={handleCancel}
          >
            취소
          </button>
        </>
      )}

      {/* 다운로드 완료 */}
      {phase === 'downloaded' && (
        <>
          <span className={`${badgeBase} bg-[#16a34a]`}>업데이트 준비됨</span>
          <span className="flex-1 min-w-0 text-[var(--text-secondary)] whitespace-nowrap overflow-hidden text-ellipsis">
            <span className="font-semibold text-[var(--accent,#1f3f62)] font-mono">{latest}</span>{' '}
            다운로드 완료
          </span>
          <button
            className={`${btnBase} bg-[var(--accent,#1f3f62)] text-white hover:bg-[var(--accent-hover,#2d5a8a)]`}
            onClick={handleApply}
          >
            적용 + 재시작
          </button>
          <button
            className={`${btnBase} bg-transparent text-[var(--text-secondary)] border border-[var(--accent-border,#c7d2fe)] hover:bg-[var(--accent-light,#eef2ff)] hover:text-[var(--text-primary)]`}
            onClick={handleDismiss}
          >
            나중에
          </button>
        </>
      )}

      {/* 적용 중 */}
      {phase === 'applying' && (
        <>
          <span className={`${badgeBase} bg-[#d97706]`}>업데이트 적용 중...</span>
          <span className="flex-1 min-w-0 text-[var(--text-muted)] italic">잠시만 기다려 주세요</span>
        </>
      )}

      {/* 재시작 필요 */}
      {phase === 'restart_required' && (
        <>
          <span className={`${badgeBase} bg-[#16a34a]`}>업데이트 완료</span>
          <span className="flex-1 min-w-0 text-[var(--text-secondary)] whitespace-nowrap overflow-hidden text-ellipsis">
            재시작하면{' '}
            <span className="font-semibold text-[var(--accent,#1f3f62)] font-mono">{latest}</span>
            이 적용됩니다.
          </span>
          {updateUrl && (
            <a
              href={updateUrl}
              target="_blank"
              rel="noopener noreferrer"
              className={`${linkBase} text-[var(--accent,#1f3f62)]`}
            >
              재시작 안내
            </a>
          )}
        </>
      )}

      {/* 에러 */}
      {phase === 'error' && (
        <>
          <span className={`${badgeBase} bg-[var(--error,#dc2626)]`}>업데이트 실패</span>
          <span className="flex-1 min-w-0 text-[var(--error,#dc2626)] whitespace-nowrap overflow-hidden text-ellipsis">
            {error || '알 수 없는 오류'}
          </span>
          <button
            className={`${btnBase} bg-transparent text-[var(--text-secondary)] border border-[var(--accent-border,#c7d2fe)] hover:bg-[var(--accent-light,#eef2ff)] hover:text-[var(--text-primary)]`}
            onClick={handleRetry}
          >
            다시 시도
          </button>
          {updateUrl && (
            <a
              href={updateUrl}
              target="_blank"
              rel="noopener noreferrer"
              className={`${linkBase} text-[var(--accent,#1f3f62)]`}
            >
              수동 다운로드
            </a>
          )}
          <button
            className="bg-transparent border-none text-lg leading-none text-[var(--text-muted)] cursor-pointer px-0.5 flex-shrink-0 ml-auto hover:text-[var(--error,#dc2626)] transition-colors"
            onClick={handleDismiss}
            title="닫기"
          >
            &times;
          </button>
        </>
      )}
    </div>
  );
}
