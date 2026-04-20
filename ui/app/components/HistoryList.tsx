"use client";

import { useState, useEffect, useCallback } from "react";
import { useRecoilValue, useSetRecoilState } from "recoil";
import { useRouter } from "next/navigation";
import { userInfoState, displayPanelOpenState, lyricsSongsState } from "@/recoilState";
import { apiClient, openDisplayWindow } from "@/lib/apiClient";
import { GenerationHistory, WorshipOrderItem, SongBlock } from "@/types";

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

const filterTabs: { key: string | undefined; label: string }[] = [
  { key: undefined, label: "전체" },
  { key: "bulletin", label: "주보" },
  { key: "lyrics_ppt", label: "가사 PPT" },
  { key: "display", label: "Display" },
];

export default function HistoryList({ open, onClose, filterType }: HistoryListProps) {
  const userInfo = useRecoilValue(userInfoState);
  const setDisplayPanelOpen = useSetRecoilState(displayPanelOpenState);
  const setLyricsSongs = useSetRecoilState(lyricsSongsState);
  const router = useRouter();
  const [items, setItems] = useState<GenerationHistory[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [activeFilter, setActiveFilter] = useState<string | undefined>(filterType);
  const [sending, setSending] = useState<number | null>(null);
  const [deleting, setDeleting] = useState<number | null>(null);

  const loadHistory = useCallback(
    (p: number, type?: string) => {
      if (!userInfo.email) return;
      setLoading(true);
      apiClient
        .getHistory(userInfo.email, type, p)
        .then((data: { items: GenerationHistory[]; page: number }) => {
          setItems(data.items || []);
          setPage(data.page || 1);
        })
        .catch(() => setItems([]))
        .finally(() => setLoading(false));
    },
    [userInfo.email]
  );

  useEffect(() => {
    if (open) {
      setActiveFilter(filterType);
      loadHistory(1, filterType);
    }
  }, [open, filterType, loadHistory]);

  const handleFilterChange = (type?: string) => {
    setActiveFilter(type);
    loadHistory(1, type);
  };

  const handleReuseLyrics = useCallback((orderData: { title: string; lyrics: string; bpm?: number }[]) => {
    const songs: SongBlock[] = orderData.map((s) => ({
      title: s.title,
      lyrics: s.lyrics,
      bpm: s.bpm ?? 100,
      expanded: true,
    }));
    setLyricsSongs(songs);
    onClose();
    router.push("/lyrics");
  }, [setLyricsSongs, onClose, router]);

  const handleDelete = useCallback(async (id: number) => {
    if (!userInfo.email) return;
    setDeleting(id);
    try {
      await apiClient.deleteHistory(id, userInfo.email);
      setItems(prev => prev.filter(item => item.id !== id));
    } catch {
      // 무시
    } finally {
      setDeleting(null);
    }
  }, [userInfo.email]);

  const handleSendToDisplay = useCallback(async (orderData: WorshipOrderItem[], historyId: number) => {
    setSending(historyId);
    try {
      setDisplayPanelOpen(true);
      openDisplayWindow();
      await apiClient.startDisplay(orderData, userInfo.english_name || "", userInfo.email, true);
      onClose();
    } catch (e) {
      console.error("Display 전송 에러:", e);
    } finally {
      setSending(null);
    }
  }, [setDisplayPanelOpen, userInfo.english_name, userInfo.email, onClose]);

  if (!open) return null;

  return (
    <div className="history_overlay" onClick={onClose}>
      <div className="history_panel" onClick={(e) => e.stopPropagation()}>
        <div className="history_header">
          <h3>생성 내역</h3>
          <button className="history_close" onClick={onClose}>
            &times;
          </button>
        </div>
        <div className="history_tabs">
          {filterTabs.map((tab) => (
            <button
              key={tab.label}
              className={`history_tab${activeFilter === tab.key ? " active" : ""}`}
              onClick={() => handleFilterChange(tab.key)}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div className="history_body" style={{ opacity: loading ? 0.5 : 1, transition: "opacity 0.15s" }}>
          {!loading && items.length === 0 ? (
            <div className="history_empty">생성 내역이 없습니다</div>
          ) : (
            items.map((item) => {
              const lyricsSongs = item.type === "lyrics_ppt" && item.order_data && Array.isArray(item.order_data) && item.order_data.length > 0
                ? (item.order_data as { title: string; lyrics: string; bpm?: number }[])
                : null;
              return (
              <div key={item.id} className="history_item">
                <div className="history_item_top">
                  <span className={`history_badge ${item.status}`}>
                    {typeLabels[item.type] || item.type}
                  </span>
                  <span className="history_date">
                    {new Date(item.created_at).toLocaleString("ko-KR")}
                  </span>
                </div>
                {item.type === "lyrics_ppt" ? (
                  lyricsSongs ? (
                    <div className="history_songs">
                      {lyricsSongs.map((s, i) => (
                        <span key={i} className="history_song_tag">{s.title}</span>
                      ))}
                    </div>
                  ) : (
                    <div className="history_filename">{item.filename}</div>
                  )
                ) : item.type === "display" && item.order_data && Array.isArray(item.order_data) && item.order_data.length > 0 ? (
                  <div className="history_songs">
                    {(item.order_data as WorshipOrderItem[]).map((o, i) => (
                      <span key={i} className="history_song_tag">{o.title}</span>
                    ))}
                  </div>
                ) : (
                  item.filename && <div className="history_filename">{item.filename}</div>
                )}
                <div className="history_item_bottom">
                  <span className={`history_status ${item.status}`}>
                    {item.status === "success" ? "성공" : "실패"}
                  </span>
                  <div style={{ display: "flex", gap: 6 }}>
                    {item.type === "lyrics_ppt" && (
                      <button
                        className="history_send_btn"
                        onClick={() => handleReuseLyrics(lyricsSongs ?? [])}
                      >
                        전송
                      </button>
                    )}
                    {item.type !== "lyrics_ppt" && item.order_data && Array.isArray(item.order_data) && item.order_data.length > 0 && (
                      <button
                        className="history_send_btn"
                        onClick={() => handleSendToDisplay(item.order_data as WorshipOrderItem[], item.id)}
                        disabled={sending === item.id}
                      >
                        {sending === item.id ? "전송 중..." : "전송"}
                      </button>
                    )}
                    <button
                      className="history_delete_btn"
                      onClick={() => handleDelete(item.id)}
                      disabled={deleting === item.id}
                      title="삭제"
                    >
                      {deleting === item.id ? "…" : "삭제"}
                    </button>
                  </div>
                </div>
              </div>
              );
            })
          )}
        </div>

        {items.length > 0 && (
          <div className="history_footer">
            <button
              disabled={page <= 1 || loading}
              onClick={() => loadHistory(page - 1, activeFilter)}
            >
              이전
            </button>
            <span>{page} 페이지</span>
            <button
              disabled={items.length < 20 || loading}
              onClick={() => loadHistory(page + 1, activeFilter)}
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
          background: var(--surface-elevated);
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
          border-bottom: 1px solid var(--border);
          flex-shrink: 0;
        }
        .history_header h3 {
          margin: 0;
          font-size: 18px;
          font-weight: 700;
          color: var(--text-primary);
        }
        .history_close {
          background: none;
          border: none;
          font-size: 24px;
          cursor: pointer;
          color: var(--text-secondary);
        }
        .history_tabs {
          display: flex;
          gap: 4px;
          padding: 12px 24px 0;
          flex-shrink: 0;
        }
        .history_tab {
          padding: 6px 14px;
          font-size: 12px;
          font-weight: 600;
          border: 1px solid var(--border);
          border-radius: 20px;
          background: var(--surface-input);
          color: var(--text-secondary);
          cursor: pointer;
          transition: all 0.15s;
        }
        .history_tab:hover {
          background: var(--surface-hover);
        }
        .history_tab.active {
          background: var(--accent);
          color: var(--surface-elevated);
          border-color: var(--accent);
        }
        .history_body {
          flex: 1;
          overflow-y: auto;
          padding: 16px 24px;
        }
        .history_loading,
        .history_empty {
          text-align: center;
          color: var(--text-muted);
          padding: 40px 0;
          font-size: 14px;
        }
        .history_item {
          padding: 12px 0;
          border-bottom: 1px solid var(--border-light);
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
          background: var(--badge-bg);
          color: var(--accent);
        }
        .history_date {
          font-size: 12px;
          color: var(--text-muted);
        }
        .history_filename {
          font-size: 13px;
          color: var(--text-primary);
          margin-top: 4px;
        }
        .history_songs {
          display: flex;
          flex-wrap: wrap;
          gap: 4px;
          margin-top: 4px;
        }
        .history_song_tag {
          font-size: 11px;
          padding: 2px 7px;
          border-radius: 10px;
          background: var(--surface-hover);
          color: var(--text-secondary);
          border: 1px solid var(--border-light);
        }
        .history_item_bottom {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-top: 4px;
        }
        .history_status {
          font-size: 11px;
        }
        .history_send_btn {
          padding: 3px 10px;
          font-size: 11px;
          font-weight: 600;
          background: var(--accent);
          color: var(--surface-elevated);
          border: none;
          border-radius: 4px;
          cursor: pointer;
        }
        .history_send_btn:hover {
          background: var(--accent-hover);
        }
        .history_send_btn:disabled {
          background: var(--text-muted);
          cursor: default;
        }
        .history_delete_btn {
          padding: 3px 10px;
          font-size: 11px;
          font-weight: 600;
          background: transparent;
          color: var(--text-muted);
          border: 1px solid var(--border-input);
          border-radius: 4px;
          cursor: pointer;
        }
        .history_delete_btn:hover {
          background: var(--error, #dc2626);
          color: white;
          border-color: var(--error, #dc2626);
        }
        .history_delete_btn:disabled {
          opacity: 0.4;
          cursor: default;
        }
        .history_status.success {
          color: var(--success);
        }
        .history_status.failed {
          color: var(--error);
        }
        .history_footer {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 12px;
          padding: 12px 24px 16px;
          border-top: 1px solid var(--border);
        }
        .history_footer button {
          padding: 6px 14px;
          font-size: 12px;
          background: var(--surface-hover);
          border: 1px solid var(--border-input);
          border-radius: 6px;
          cursor: pointer;
        }
        .history_footer button:disabled {
          opacity: 0.4;
          cursor: default;
        }
        .history_footer span {
          font-size: 13px;
          color: var(--text-secondary);
        }
      `}</style>
    </div>
  );
}
