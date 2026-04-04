"use client";

import { useState, useEffect, useCallback } from "react";
import { useRecoilValue } from "recoil";
import { userInfoState } from "@/recoilState";
import { apiClient } from "@/lib/apiClient";
import { GenerationHistory } from "@/types";

interface HistoryListProps {
  open: boolean;
  onClose: () => void;
  filterType?: string;
}

const typeLabels: Record<string, string> = {
  bulletin: "주보",
  ppt: "PPT",
  lyrics_ppt: "가사 PPT",
  display: "Display",
};

export default function HistoryList({ open, onClose, filterType }: HistoryListProps) {
  const userInfo = useRecoilValue(userInfoState);
  const [items, setItems] = useState<GenerationHistory[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);

  const loadHistory = useCallback(
    (p: number) => {
      if (!userInfo.email) return;
      setLoading(true);
      apiClient
        .getHistory(userInfo.email, filterType, p)
        .then((data: { items: GenerationHistory[]; page: number }) => {
          setItems(data.items || []);
          setPage(data.page || 1);
        })
        .catch(() => setItems([]))
        .finally(() => setLoading(false));
    },
    [userInfo.email, filterType]
  );

  useEffect(() => {
    if (open) loadHistory(1);
  }, [open, loadHistory]);

  if (!open) return null;

  return (
    <div className="history_overlay" onClick={onClose}>
      <div className="history_panel" onClick={(e) => e.stopPropagation()}>
        <div className="history_header">
          <h3>{filterType ? `${typeLabels[filterType] || filterType} 생성 내역` : "전체 생성 내역"}</h3>
          <button className="history_close" onClick={onClose}>
            &times;
          </button>
        </div>

        <div className="history_body">
          {loading ? (
            <div className="history_loading">불러오는 중...</div>
          ) : items.length === 0 ? (
            <div className="history_empty">생성 내역이 없습니다</div>
          ) : (
            items.map((item) => (
              <div key={item.id} className="history_item">
                <div className="history_item_top">
                  <span className={`history_badge ${item.status}`}>
                    {typeLabels[item.type] || item.type}
                  </span>
                  <span className="history_date">
                    {new Date(item.created_at).toLocaleString("ko-KR")}
                  </span>
                </div>
                {item.filename && (
                  <div className="history_filename">{item.filename}</div>
                )}
                <div className={`history_status ${item.status}`}>
                  {item.status === "success" ? "성공" : "실패"}
                </div>
              </div>
            ))
          )}
        </div>

        {items.length > 0 && (
          <div className="history_footer">
            <button
              disabled={page <= 1 || loading}
              onClick={() => loadHistory(page - 1)}
            >
              이전
            </button>
            <span>{page} 페이지</span>
            <button
              disabled={items.length < 20 || loading}
              onClick={() => loadHistory(page + 1)}
            >
              다음
            </button>
          </div>
        )}
      </div>

      <style jsx>{`
        .history_overlay {
          position: fixed;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          background: rgba(0, 0, 0, 0.5);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 11000;
        }
        .history_panel {
          background: #fff;
          border-radius: 16px;
          width: 500px;
          max-width: 90vw;
          max-height: 80vh;
          display: flex;
          flex-direction: column;
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
        }
        .history_header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 24px 16px;
          border-bottom: 1px solid #e5e7eb;
          flex-shrink: 0;
        }
        .history_header h3 {
          margin: 0;
          font-size: 18px;
          font-weight: 700;
          color: #1f2937;
        }
        .history_close {
          background: none;
          border: none;
          font-size: 24px;
          cursor: pointer;
          color: #6b7280;
        }
        .history_body {
          flex: 1;
          overflow-y: auto;
          padding: 16px 24px;
        }
        .history_loading,
        .history_empty {
          text-align: center;
          color: #9ca3af;
          padding: 40px 0;
          font-size: 14px;
        }
        .history_item {
          padding: 12px 0;
          border-bottom: 1px solid #f1f5f9;
        }
        .history_item_top {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 4px;
        }
        .history_badge {
          font-size: 11px;
          font-weight: 700;
          padding: 2px 8px;
          border-radius: 4px;
          background: #eef2ff;
          color: #1f3f62;
        }
        .history_date {
          font-size: 12px;
          color: #9ca3af;
        }
        .history_filename {
          font-size: 13px;
          color: #374151;
          margin-top: 4px;
        }
        .history_status {
          font-size: 11px;
          margin-top: 4px;
        }
        .history_status.success {
          color: #059669;
        }
        .history_status.failed {
          color: #dc2626;
        }
        .history_footer {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 12px;
          padding: 12px 24px 16px;
          border-top: 1px solid #e5e7eb;
        }
        .history_footer button {
          padding: 6px 14px;
          font-size: 12px;
          background: #f3f4f6;
          border: 1px solid #d1d5db;
          border-radius: 6px;
          cursor: pointer;
        }
        .history_footer button:disabled {
          opacity: 0.4;
          cursor: default;
        }
        .history_footer span {
          font-size: 13px;
          color: #6b7280;
        }
      `}</style>
    </div>
  );
}
