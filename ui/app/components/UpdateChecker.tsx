"use client";

import { useEffect, useState, useCallback } from "react";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ||
  (typeof window !== "undefined" ? window.location.origin : "http://localhost:8080");

interface UpdateInfo {
  ok: boolean;
  current: string;
  latest?: string;
  updateUrl?: string;
  notes?: string;
  hasUpdate?: boolean;
  error?: string;
}

const DISMISS_KEY = "update_dismissed_version";
const CHECK_INTERVAL_MS = 30 * 60 * 1000; // 30분

export default function UpdateChecker() {
  const [info, setInfo] = useState<UpdateInfo | null>(null);
  const [dismissed, setDismissed] = useState(false);

  const checkUpdate = useCallback(async () => {
    try {
      const res = await fetch(`${BASE_URL}/api/update/check`);
      if (!res.ok) return;
      const data: UpdateInfo = await res.json();
      setInfo(data);

      // 이미 이 버전으로 닫은 적 있으면 dismissed 처리
      if (data.latest) {
        const saved = localStorage.getItem(DISMISS_KEY);
        if (saved === data.latest) {
          setDismissed(true);
        }
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

  const handleDismiss = () => {
    if (info?.latest) {
      localStorage.setItem(DISMISS_KEY, info.latest);
    }
    setDismissed(true);
  };

  // 업데이트 없음, 오류, 닫힘 → 렌더 없음
  if (!info || !info.hasUpdate || dismissed) return null;

  return (
    <div className="update_banner">
      <span className="update_badge">업데이트 가능</span>
      <span className="update_text">
        현재 버전{" "}
        <span className="update_ver">{info.current}</span>
        {" "}→{" "}
        <span className="update_ver latest">{info.latest}</span>
      </span>
      {info.updateUrl && (
        <a
          href={info.updateUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="update_link"
        >
          릴리스 노트 보기
        </a>
      )}
      <button className="update_dismiss" onClick={handleDismiss} title="닫기">
        &times;
      </button>

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
        }

        .update_text {
          flex: 1;
          min-width: 0;
          color: var(--text-secondary);
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .update_ver {
          font-weight: 600;
          color: var(--text-primary);
          font-family: monospace;
          font-size: 13px;
        }

        .update_ver.latest {
          color: var(--accent);
        }

        .update_link {
          padding: 3px 12px;
          border-radius: 6px;
          background: var(--accent);
          color: #fff !important;
          font-size: 12px;
          font-weight: 600;
          text-decoration: none;
          flex-shrink: 0;
          transition: background 0.15s;
        }

        .update_link:hover {
          background: var(--accent-hover, #2d5a8a);
        }

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
        }
      `}</style>
    </div>
  );
}
