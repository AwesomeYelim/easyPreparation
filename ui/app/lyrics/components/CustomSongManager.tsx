"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import toast from "react-hot-toast";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ||
  (typeof window !== "undefined" ? window.location.origin : "http://localhost:8080");

interface Song {
  id: number;
  title: string;
  artist: string;
  lyrics: string;
  tags: string[];
  used_count: number;
  created_at: string;
  updated_at: string;
}

interface SongForm {
  title: string;
  artist: string;
  lyrics: string;
  tags: string;
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, init);
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(text || `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

const songsApi = {
  list: () => apiFetch<Song[]>("/api/songs/"),
  search: (q: string) => apiFetch<Song[]>(`/api/songs/search?q=${encodeURIComponent(q)}`),
  create: (data: { title: string; artist?: string; lyrics: string; tags?: string[] }) =>
    apiFetch<Song>("/api/songs/", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    }),
  update: (id: number, data: { title: string; artist?: string; lyrics: string; tags?: string[] }) =>
    apiFetch<Song>(`/api/songs/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    }),
  remove: (id: number) =>
    apiFetch<{ ok: boolean }>(`/api/songs/${id}`, { method: "DELETE" }),
};

function emptyForm(): SongForm {
  return { title: "", artist: "", lyrics: "", tags: "" };
}

function songToForm(s: Song): SongForm {
  return { title: s.title, artist: s.artist, lyrics: s.lyrics, tags: (s.tags ?? []).join(", ") };
}

function formToPayload(f: SongForm) {
  return {
    title: f.title.trim(),
    artist: f.artist.trim() || undefined,
    lyrics: f.lyrics,
    tags: f.tags.split(",").map((t) => t.trim()).filter(Boolean),
  };
}

export default function CustomSongManager() {
  const [songs, setSongs] = useState<Song[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState<Song | null>(null);
  const [form, setForm] = useState<SongForm>(emptyForm());
  const [isNew, setIsNew] = useState(false);
  const [saving, setSaving] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const loadList = useCallback(async () => {
    setLoading(true);
    try {
      const data = await songsApi.list();
      setSongs(data ?? []);
    } catch {
      toast.error("목록을 불러오지 못했습니다.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadList(); }, [loadList]);

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(async () => {
      if (!query.trim()) { loadList(); return; }
      setLoading(true);
      try {
        const data = await songsApi.search(query);
        setSongs(data ?? []);
      } catch {
        toast.error("검색에 실패했습니다.");
      } finally {
        setLoading(false);
      }
    }, 300);
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current); };
  }, [query, loadList]);

  const selectSong = (s: Song) => { setSelected(s); setForm(songToForm(s)); setIsNew(false); };
  const handleNew = () => { setSelected(null); setForm(emptyForm()); setIsNew(true); };

  const handleSave = async () => {
    if (!form.title.trim()) { toast.error("제목을 입력해주세요."); return; }
    setSaving(true);
    try {
      const payload = formToPayload(form);
      if (isNew) {
        const created = await songsApi.create(payload);
        toast.success("곡이 추가되었습니다.");
        await loadList();
        setSelected(created);
        setForm(songToForm(created));
        setIsNew(false);
      } else if (selected) {
        const updated = await songsApi.update(selected.id, payload);
        toast.success("저장되었습니다.");
        setSongs((prev) => prev.map((s) => (s.id === updated.id ? updated : s)));
        setSelected(updated);
        setForm(songToForm(updated));
      }
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "저장 실패");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!selected) return;
    if (!window.confirm(`"${selected.title}"을(를) 삭제하시겠습니까?`)) return;
    try {
      await songsApi.remove(selected.id);
      toast.success("삭제되었습니다.");
      setSongs((prev) => prev.filter((s) => s.id !== selected.id));
      setSelected(null);
      setForm(emptyForm());
      setIsNew(false);
    } catch (e: unknown) {
      toast.error(e instanceof Error ? e.message : "삭제 실패");
    }
  };

  const setField = <K extends keyof SongForm>(key: K, value: SongForm[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }));

  const hasChanges =
    (isNew && (form.title || form.lyrics)) ||
    (!isNew && selected && (
      form.title !== selected.title ||
      form.artist !== (selected.artist ?? "") ||
      form.lyrics !== selected.lyrics ||
      form.tags !== (selected.tags ?? []).join(", ")
    ));

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* 헤더 */}
      <div className="flex items-center gap-3 px-4 py-3 border-b border-pro-border shrink-0">
        <span
          className="material-symbols-outlined text-pro-accent"
          style={{ fontSize: "18px", fontVariationSettings: "'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24" }}
        >
          library_music
        </span>
        <span className="text-xs font-semibold text-pro-text">내 찬양 라이브러리</span>
        <p className="text-[10px] text-pro-text-muted ml-1">가사 자동 검색 시 1순위로 참조됩니다</p>

        {/* 검색창 */}
        <div className="relative ml-2" style={{ width: 200 }}>
          <span
            className="material-symbols-outlined absolute left-2 top-1/2 -translate-y-1/2 text-pro-text-muted pointer-events-none"
            style={{ fontSize: "14px" }}
          >
            search
          </span>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="제목, 아티스트, 가사..."
            className="w-full bg-white/5 border border-pro-border rounded-lg pl-7 pr-3 py-1.5 text-xs text-pro-text placeholder-pro-text-muted focus:outline-none focus:border-pro-accent transition-colors"
          />
        </div>

        <div className="flex items-center gap-2 ml-auto">
          <button
            onClick={handleNew}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs text-white bg-pro-accent hover:bg-pro-accent/80 transition-colors"
          >
            <span className="material-symbols-outlined" style={{ fontSize: "13px" }}>add</span>
            새 곡
          </button>
        </div>
      </div>

      {/* 본문 */}
      <div className="flex flex-1 overflow-hidden">
        {/* 목록 */}
        <div className="w-56 shrink-0 flex flex-col border-r border-pro-border overflow-hidden">
          <div className="px-3 py-2 shrink-0">
            <span className="text-[10px] text-pro-text-muted">
              {loading ? "불러오는 중..." : `${songs.length}곡`}
            </span>
          </div>
          <div className="flex-1 overflow-y-auto">
            {loading ? (
              <div className="flex items-center justify-center h-20">
                <span className="text-xs text-pro-text-muted">로딩 중...</span>
              </div>
            ) : songs.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-32 gap-2 text-pro-text-muted">
                <span className="material-symbols-outlined" style={{ fontSize: "28px" }}>music_note</span>
                <span className="text-xs">{query ? "검색 결과 없음" : "곡이 없습니다"}</span>
              </div>
            ) : (
              <ul>
                {songs.map((s) => {
                  const isActive = selected?.id === s.id && !isNew;
                  return (
                    <li key={s.id}>
                      <button
                        onClick={() => selectSong(s)}
                        className={`w-full text-left px-3 py-2.5 transition-colors border-l-2 ${
                          isActive ? "bg-pro-accent/15 border-pro-accent" : "border-transparent hover:bg-white/5"
                        }`}
                      >
                        <div className="text-xs font-medium text-pro-text truncate">{s.title}</div>
                        <div className="flex items-center gap-2 mt-0.5">
                          {s.artist && <span className="text-[11px] text-pro-text-muted truncate">{s.artist}</span>}
                          {s.used_count > 0 && (
                            <span className="ml-auto text-[10px] text-pro-text-muted shrink-0">{s.used_count}회</span>
                          )}
                        </div>
                        {s.tags && s.tags.length > 0 && (
                          <div className="flex flex-wrap gap-1 mt-1">
                            {s.tags.slice(0, 2).map((tag) => (
                              <span key={tag} className="text-[10px] bg-pro-accent/10 text-pro-accent px-1.5 py-0.5 rounded">
                                {tag}
                              </span>
                            ))}
                          </div>
                        )}
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>
        </div>

        {/* 편집 */}
        <div className="flex-1 flex flex-col overflow-hidden">
          {!selected && !isNew ? (
            <div className="flex flex-col items-center justify-center h-full gap-3 text-pro-text-muted">
              <span
                className="material-symbols-outlined"
                style={{ fontSize: "40px", fontVariationSettings: "'FILL' 0, 'wght' 200, 'GRAD' 0, 'opsz' 48" }}
              >library_music</span>
              <p className="text-sm">좌측에서 곡을 선택하거나</p>
              <button onClick={handleNew} className="text-pro-accent text-sm hover:underline">새 곡을 추가하세요</button>
            </div>
          ) : (
            <>
              <div className="flex items-center justify-between px-4 py-3 border-b border-pro-border shrink-0">
                <span className="text-xs font-medium text-pro-text">{isNew ? "새 곡 추가" : "곡 편집"}</span>
                <div className="flex items-center gap-2">
                  {!isNew && selected && (
                    <button
                      onClick={handleDelete}
                      className="flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs text-red-400 hover:bg-red-400/10 transition-colors"
                    >
                      <span className="material-symbols-outlined" style={{ fontSize: "13px" }}>delete</span>삭제
                    </button>
                  )}
                  <button
                    onClick={handleSave}
                    disabled={saving || !hasChanges}
                    className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs text-white bg-pro-accent hover:bg-pro-accent/80 disabled:opacity-40 transition-colors"
                  >
                    <span className="material-symbols-outlined" style={{ fontSize: "13px" }}>save</span>
                    {saving ? "저장 중..." : "저장"}
                  </button>
                </div>
              </div>
              <div className="flex-1 overflow-y-auto p-4 space-y-4">
                <div>
                  <label className="block text-xs text-pro-text-muted mb-1">제목 <span className="text-red-400">*</span></label>
                  <input
                    type="text"
                    value={form.title}
                    onChange={(e) => setField("title", e.target.value)}
                    placeholder="곡 제목"
                    className="w-full bg-white/5 border border-pro-border rounded-lg px-3 py-2 text-sm text-pro-text placeholder-pro-text-muted focus:outline-none focus:border-pro-accent transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs text-pro-text-muted mb-1">아티스트 / 작곡가</label>
                  <input
                    type="text"
                    value={form.artist}
                    onChange={(e) => setField("artist", e.target.value)}
                    placeholder="아티스트 (선택)"
                    className="w-full bg-white/5 border border-pro-border rounded-lg px-3 py-2 text-sm text-pro-text placeholder-pro-text-muted focus:outline-none focus:border-pro-accent transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs text-pro-text-muted mb-1">
                    가사
                    <span className="ml-2 text-[10px] opacity-60">(절 구분: <code className="bg-white/10 px-1 rounded">---</code>)</span>
                  </label>
                  <textarea
                    value={form.lyrics}
                    onChange={(e) => setField("lyrics", e.target.value)}
                    placeholder={"1절\n가사를 입력하세요\n---\n2절\n가사를 입력하세요"}
                    rows={12}
                    className="w-full bg-white/5 border border-pro-border rounded-lg px-3 py-2 text-sm text-pro-text placeholder-pro-text-muted focus:outline-none focus:border-pro-accent font-mono resize-none transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs text-pro-text-muted mb-1">
                    태그 <span className="ml-1 text-[10px] opacity-60">(쉼표로 구분)</span>
                  </label>
                  <input
                    type="text"
                    value={form.tags}
                    onChange={(e) => setField("tags", e.target.value)}
                    placeholder="예: 찬양, 경배, 주일"
                    className="w-full bg-white/5 border border-pro-border rounded-lg px-3 py-2 text-sm text-pro-text placeholder-pro-text-muted focus:outline-none focus:border-pro-accent transition-colors"
                  />
                  {form.tags.trim() && (
                    <div className="flex flex-wrap gap-1 mt-2">
                      {form.tags.split(",").map((t) => t.trim()).filter(Boolean).map((tag) => (
                        <span key={tag} className="text-[11px] bg-pro-accent/10 text-pro-accent px-2 py-0.5 rounded-full">{tag}</span>
                      ))}
                    </div>
                  )}
                </div>
                {!isNew && selected && (
                  <div className="text-[10px] text-pro-text-muted space-y-0.5 pt-2 border-t border-pro-border">
                    <div>사용 횟수: {selected.used_count}회</div>
                    <div>생성: {new Date(selected.created_at).toLocaleDateString("ko-KR", { year: "numeric", month: "short", day: "numeric" })}</div>
                    <div>수정: {new Date(selected.updated_at).toLocaleDateString("ko-KR", { year: "numeric", month: "short", day: "numeric" })}</div>
                  </div>
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
