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

  // 업데이트 체크
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

  // 30분마다 자동 체크
  useEffect(() => {
    checkUpdate();
    const timer = setInterval(checkUpdate, CHECK_INTERVAL_MS);
    return () => clearInterval(timer);
  }, [checkUpdate]);

  // WS 구독: update_progress 메시지 수신
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

  // WS가 없을 때 폴링 fallback
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

  // 핸들러
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
    if (state.latest) {
      localStorage.setItem(DISMISS_KEY, state.latest);
    }
    setDismissed(true);
  };

  const handleRetry = () => {
    setState(s => ({ ...s, phase: 'idle', error: '' }));
    checkUpdate();
  };

  // 렌더 조건: 알림 없고 idle이거나, 알림 닫힌 상태면 숨김
  // 다운로드/적용 등 진행 중인 phase는 항상 표시
  if (state.phase === 'idle' && (!state.hasUpdate || dismissed)) return null;

  const { phase, percent, totalBytes, downloadedBytes, latest, current, updateUrl, error } = state;

  return (
    <div className={`update_banner ${phase !== 'idle' ? 'update_banner--active' : ''}`}>

      {/* 알림 상태 */}
      {phase === 'idle' && state.hasUpdate && (
        <>
          <span className="update_badge">업데이트 가능</span>
          <span className="update_text">
            현재 버전 <span className="update_ver">{current}</span>
            {' '}→{' '}
            <span className="update_ver latest">{latest}</span>
          </span>
          <button className="update_btn update_btn--primary" onClick={handleStartDownload}>
            지금 업데이트
          </button>
          {updateUrl && (
            <a
              href={updateUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="update_link"
            >
              릴리스 노트
            </a>
          )}
          <button className="update_dismiss" onClick={handleDismiss} title="닫기">
            &times;
          </button>
        </>
      )}

      {/* 다운로드 중 */}
      {phase === 'downloading' && (
        <>
          <span className="update_badge update_badge--progress">다운로드 중...</span>
          <div className="update_progress_wrap">
            <div className="update_progress_bar">
              <div className="update_progress_fill" style={{ width: `${percent}%` }} />
            </div>
            <span className="update_progress_text">
              {totalBytes > 0
                ? `${formatBytes(downloadedBytes)} / ${formatBytes(totalBytes)} (${Math.round(percent)}%)`
                : `${Math.round(percent)}%`}
            </span>
          </div>
          <button className="update_btn update_btn--ghost" onClick={handleCancel}>
            취소
          </button>
        </>
      )}

      {/* 다운로드 완료 */}
      {phase === 'downloaded' && (
        <>
          <span className="update_badge update_badge--success">업데이트 준비됨</span>
          <span className="update_text">
            <span className="update_ver latest">{latest}</span> 다운로드 완료
          </span>
          <button className="update_btn update_btn--primary" onClick={handleApply}>
            적용 + 재시작
          </button>
          <button className="update_btn update_btn--ghost" onClick={handleDismiss}>
            나중에
          </button>
        </>
      )}

      {/* 적용 중 */}
      {phase === 'applying' && (
        <>
          <span className="update_badge update_badge--progress">업데이트 적용 중...</span>
          <span className="update_text update_text--muted">잠시만 기다려 주세요</span>
        </>
      )}

      {/* 재시작 필요 */}
      {phase === 'restart_required' && (
        <>
          <span className="update_badge update_badge--success">업데이트 완료</span>
          <span className="update_text">
            재시작하면 <span className="update_ver latest">{latest}</span>이 적용됩니다.
          </span>
          {updateUrl && (
            <a
              href={updateUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="update_link"
            >
              재시작 안내
            </a>
          )}
        </>
      )}

      {/* 에러 */}
      {phase === 'error' && (
        <>
          <span className="update_badge update_badge--error">업데이트 실패</span>
          <span className="update_text update_text--error">{error || '알 수 없는 오류'}</span>
          <button className="update_btn update_btn--ghost" onClick={handleRetry}>
            다시 시도
          </button>
          {updateUrl && (
            <a
              href={updateUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="update_link"
            >
              수동 다운로드
            </a>
          )}
          <button className="update_dismiss" onClick={handleDismiss} title="닫기">
            &times;
          </button>
        </>
      )}

      <style jsx>{`
        .update_banner {
          width: 100%;
          display: flex;
          align-items: center;
          gap: 10px;
          padding: 8px 20px;
          background: var(--accent-light, #eef2ff);
          border-bottom: 1px solid var(--accent-border, #c7d2fe);
          font-size: 13px;
          color: var(--text-primary);
          box-sizing: border-box;
          flex-wrap: wrap;
        }

        .update_banner--active {
          background: var(--surface-elevated, #f8fafc);
          border-bottom-color: var(--accent-border, #c7d2fe);
        }

        /* 배지 */
        .update_badge {
          display: inline-flex;
          align-items: center;
          padding: 2px 8px;
          border-radius: 10px;
          background: var(--accent, #1f3f62);
          color: #fff;
          font-size: 11px;
          font-weight: 700;
          letter-spacing: 0.3px;
          flex-shrink: 0;
          white-space: nowrap;
        }

        .update_badge--progress {
          background: #d97706;
        }

        .update_badge--success {
          background: #16a34a;
        }

        .update_badge--error {
          background: var(--error, #dc2626);
        }

        /* 텍스트 */
        .update_text {
          flex: 1;
          min-width: 0;
          color: var(--text-secondary);
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .update_text--muted {
          color: var(--text-muted);
          font-style: italic;
        }

        .update_text--error {
          color: var(--error, #dc2626);
        }

        .update_ver {
          font-weight: 600;
          color: var(--text-primary);
          font-family: monospace;
          font-size: 13px;
        }

        .update_ver.latest {
          color: var(--accent, #1f3f62);
        }

        /* 프로그레스 바 */
        .update_progress_wrap {
          flex: 1;
          display: flex;
          align-items: center;
          gap: 10px;
          min-width: 0;
        }

        .update_progress_bar {
          flex: 1;
          height: 6px;
          background: var(--surface-input, #e2e8f0);
          border-radius: 3px;
          overflow: hidden;
          min-width: 80px;
        }

        .update_progress_fill {
          height: 100%;
          background: var(--accent, #1f3f62);
          border-radius: 3px;
          transition: width 0.3s ease;
        }

        .update_progress_text {
          font-size: 12px;
          color: var(--text-secondary);
          white-space: nowrap;
          flex-shrink: 0;
          font-variant-numeric: tabular-nums;
        }

        /* 버튼 */
        .update_btn {
          padding: 3px 12px;
          border-radius: 6px;
          font-size: 12px;
          font-weight: 600;
          cursor: pointer;
          flex-shrink: 0;
          border: none;
          transition: background 0.15s, color 0.15s;
          white-space: nowrap;
        }

        .update_btn--primary {
          background: var(--accent, #1f3f62);
          color: #fff;
        }

        .update_btn--primary:hover {
          background: var(--accent-hover, #2d5a8a);
        }

        .update_btn--ghost {
          background: transparent;
          color: var(--text-secondary);
          border: 1px solid var(--accent-border, #c7d2fe);
        }

        .update_btn--ghost:hover {
          background: var(--accent-light, #eef2ff);
          color: var(--text-primary);
        }

        /* 릴리스 노트 링크 */
        .update_link {
          padding: 3px 12px;
          border-radius: 6px;
          background: transparent;
          color: var(--accent, #1f3f62) !important;
          font-size: 12px;
          font-weight: 600;
          text-decoration: none;
          flex-shrink: 0;
          border: 1px solid var(--accent-border, #c7d2fe);
          transition: background 0.15s;
          white-space: nowrap;
        }

        .update_link:hover {
          background: var(--accent-light, #eef2ff);
        }

        /* 닫기 버튼 */
        .update_dismiss {
          background: none;
          border: none;
          font-size: 18px;
          line-height: 1;
          color: var(--text-muted);
          cursor: pointer;
          padding: 0 2px;
          flex-shrink: 0;
          margin-left: auto;
          transition: color 0.15s;
        }

        .update_dismiss:hover {
          color: var(--error, #dc2626);
        }

        @media (max-width: 600px) {
          .update_banner {
            padding: 8px 12px;
            gap: 8px;
          }
          .update_text {
            white-space: normal;
          }
          .update_progress_wrap {
            flex-direction: column;
            align-items: flex-start;
            gap: 4px;
          }
          .update_progress_bar {
            width: 100%;
          }
        }
      `}</style>
    </div>
  );
}
