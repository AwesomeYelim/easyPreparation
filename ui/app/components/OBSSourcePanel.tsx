"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { apiClient } from "@/lib/apiClient";
import { OBSSourceItem, OBSDevice, OBSInitialSetupResult } from "@/types";

interface OBSSourcePanelProps {
  open: boolean;
  onClose: () => void;
}

type Tab = "setup" | "logo" | "camera" | "display" | "sources";

const TABS: { key: Tab; label: string }[] = [
  { key: "setup", label: "초기 설정" },
  { key: "logo", label: "로고" },
  { key: "camera", label: "카메라" },
  { key: "display", label: "Display" },
  { key: "sources", label: "전체 소스" },
];

const POSITIONS = [
  { key: "top-left", label: "좌상" },
  { key: "top-right", label: "우상" },
  { key: "bottom-left", label: "좌하" },
  { key: "bottom-right", label: "우하" },
];

export default function OBSSourcePanel({ open, onClose }: OBSSourcePanelProps) {
  const [tab, setTab] = useState<Tab>("logo");
  const [scenes, setScenes] = useState<string[]>([]);
  const [connected, setConnected] = useState(false);
  const [selectedScene, setSelectedScene] = useState("");
  const [toastMsg, setToastMsg] = useState<{ msg: string; type: "error" | "info" } | null>(null);

  // Logo state
  const [logoUploaded, setLogoUploaded] = useState(false);
  const [logoPosition, setLogoPosition] = useState("top-right");
  const [logoScale, setLogoScale] = useState(0.15);
  const [logoApplied, setLogoApplied] = useState(false);
  const [logoItemId, setLogoItemId] = useState<number | null>(null);
  const [dragOver, setDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Camera state
  const [devices, setDevices] = useState<OBSDevice[]>([]);
  const [selectedDevice, setSelectedDevice] = useState("");
  const [cameraName, setCameraName] = useState("카메라");
  const [loadingDevices, setLoadingDevices] = useState(false);

  // Sources state
  const [sources, setSources] = useState<OBSSourceItem[]>([]);
  const [loadingSources, setLoadingSources] = useState(false);

  // Display setup state
  const [displayURL, setDisplayURL] = useState("http://localhost:8080/display");
  const [displaySetupDone, setDisplaySetupDone] = useState(false);

  // Connect form state
  const [obsIP, setObsIP] = useState("localhost");
  const [obsPort, setObsPort] = useState(4455);
  const [obsPassword, setObsPassword] = useState("");
  const [connectLoading, setConnectLoading] = useState(false);

  const [busy, setBusy] = useState(false);
  const [setupLoading, setSetupLoading] = useState(false);
  const [setupResult, setSetupResult] = useState<OBSInitialSetupResult | null>(null);
  const [setupDone, setSetupDone] = useState(false);

  const showToast = (msg: string, type: "error" | "info" = "error") => {
    setToastMsg({ msg, type });
    setTimeout(() => setToastMsg(null), 3000);
  };

  const fetchScenes = useCallback(async () => {
    try {
      // 연결 상태는 feature gate 없는 /obs/status로 확인
      const status = await apiClient.obsGetStatus();
      setConnected(status.connected);
      if (status.currentScene && !selectedScene) {
        setSelectedScene(status.currentScene);
      }
      // 씬 목록은 Pro 기능 — 실패해도 connected 상태는 유지
      const res = await apiClient.getOBSScenes();
      if (!res.error) {
        setScenes(res.scenes || []);
      }
    } catch {
      setConnected(false);
      setScenes([]);
    }
  }, [selectedScene]);

  const fetchSources = useCallback(async () => {
    if (!selectedScene) return;
    setLoadingSources(true);
    try {
      const res = await apiClient.getOBSSources(selectedScene);
      if ((res as any).error) {
        // feature_locked 등 에러 응답 — 조용히 무시
        setLoadingSources(false);
        return;
      }
      setSources(res.items || []);
      const logoItem = (res.items || []).find((s) => s.sourceName === "EP_Logo");
      if (logoItem) {
        setLogoApplied(true);
        setLogoItemId(logoItem.sceneItemId);
      }
    } catch {
      setSources([]);
    }
    setLoadingSources(false);
  }, [selectedScene]);

  useEffect(() => {
    if (open) fetchScenes();
  }, [open, fetchScenes]);

  useEffect(() => {
    if (open && selectedScene) fetchSources();
  }, [open, selectedScene, fetchSources]);

  const handleConnect = async () => {
    setConnectLoading(true);
    try {
      const res = await apiClient.obsConnect(obsIP, obsPort, obsPassword);
      if (res.connected) {
        setConnected(true);
        showToast("OBS 연결 완료", "info");
        fetchScenes();
        if (tab === "setup") fetchDevices();
      } else {
        setConnected(false);
        showToast("연결 실패 — IP/포트/비밀번호를 확인하세요");
      }
    } catch {
      showToast("연결 실패");
    }
    setConnectLoading(false);
  };

  const handleSetupInitial = async () => {
    setSetupLoading(true);
    setSetupResult(null);
    try {
      const res = await apiClient.obsSetupInitial(selectedDevice || undefined);
      if ((res as any).error) {
        showToast((res as any).message || "초기 설정 실패");
        setSetupLoading(false);
        return;
      }
      setSetupResult(res);
      if (res.success) {
        setSetupDone(true);
        const msg =
          (res.warnings ?? []).length > 0
            ? `초기 설정 완료 (경고 ${res.warnings.length}건)`
            : "초기 설정 완료";
        showToast(msg, "info");
        fetchScenes();
        fetchSources();
      } else {
        showToast("초기 설정 실패");
      }
    } catch {
      showToast("초기 설정 실패");
    }
    setSetupLoading(false);
  };

  const handleLogoUpload = async (fileList: FileList | null) => {
    if (!fileList || fileList.length === 0) return;
    const file = fileList[0];
    const ext = file.name.toLowerCase().split(".").pop();
    if (!["png", "jpg", "jpeg"].includes(ext || "")) {
      showToast("PNG/JPG 파일만 업로드 가능합니다.");
      return;
    }
    setBusy(true);
    try {
      const res = await apiClient.uploadOBSLogo(file);
      if (res.ok) {
        setLogoUploaded(true);
        showToast("로고 업로드 완료", "info");
      }
    } catch {
      showToast("업로드 실패");
    }
    setBusy(false);
  };

  const handleLogoApply = async () => {
    if (!selectedScene) {
      showToast("씬을 선택하세요");
      return;
    }
    setBusy(true);
    try {
      const res = await apiClient.applyOBSLogo(selectedScene, logoPosition, logoScale);
      if (res.ok) {
        setLogoApplied(true);
        setLogoItemId(res.sceneItemId ?? null);
        showToast("OBS에 로고 적용 완료", "info");
        fetchSources();
      } else {
        showToast(res.error || "적용 실패");
      }
    } catch {
      showToast("적용 실패");
    }
    setBusy(false);
  };

  const fetchDevices = async () => {
    setLoadingDevices(true);
    try {
      const res = await apiClient.getOBSCameraDevices();
      setDevices(res.devices || []);
      if (res.devices?.length && !selectedDevice) {
        setSelectedDevice(res.devices[0].value);
      }
    } catch {
      setDevices([]);
    }
    setLoadingDevices(false);
  };

  const handleAddCamera = async () => {
    if (!selectedScene || !selectedDevice) {
      showToast("씬과 카메라를 선택하세요");
      return;
    }
    setBusy(true);
    try {
      const res = await apiClient.addOBSCamera(selectedScene, selectedDevice, cameraName || "카메라");
      if (res.ok) {
        showToast("카메라 추가 완료", "info");
        fetchSources();
      } else {
        showToast(res.error || "추가 실패");
      }
    } catch {
      showToast("추가 실패");
    }
    setBusy(false);
  };

  const handleSetupDisplay = async () => {
    if (!selectedScene) {
      showToast("씬을 선택하세요");
      return;
    }
    setBusy(true);
    try {
      const res = await apiClient.setupOBSDisplay(selectedScene, displayURL);
      if (res.ok) {
        setDisplaySetupDone(true);
        showToast("Display 소스 설정 완료", "info");
        fetchSources();
      } else {
        showToast(res.error || "설정 실패");
      }
    } catch {
      showToast("설정 실패");
    }
    setBusy(false);
  };

  const handleToggle = async (item: OBSSourceItem) => {
    try {
      await apiClient.toggleOBSSource(selectedScene, item.sceneItemId, !item.enabled);
      fetchSources();
    } catch {
      showToast("토글 실패");
    }
  };

  const handleRemove = async (item: OBSSourceItem) => {
    try {
      const res = await apiClient.removeOBSSource(item.sourceName);
      if (res.ok) {
        if (item.sourceName === "EP_Logo") {
          setLogoApplied(false);
          setLogoItemId(null);
        }
        fetchSources();
      } else {
        showToast(res.error || "삭제 실패");
      }
    } catch {
      showToast("삭제 실패");
    }
  };

  if (!open) return null;

  const selectClass =
    "w-full px-2.5 py-2 bg-white/10 border border-white/20 rounded-md text-white text-xs outline-none";
  const btnPrimaryClass = `px-5 py-2 bg-[#204d87] border-none rounded-md text-white text-xs font-semibold transition-opacity ${
    busy ? "opacity-60 cursor-default" : "cursor-pointer hover:bg-[#2d5a8a]"
  }`;

  return (
    <div className="fixed inset-0 z-[10600] flex items-center justify-center">
      {/* overlay */}
      <div className="absolute inset-0 bg-black/60" onClick={onClose} />

      {/* modal */}
      <div className="relative w-[640px] max-h-[80vh] bg-[#2c2c2c] rounded-xl flex flex-col overflow-hidden">
        {/* header */}
        <div className="flex justify-between items-center px-5 py-4 border-b border-white/10">
          <h3 className="m-0 text-white text-base font-semibold">OBS 소스 관리</h3>
          <div className="flex items-center gap-3">
            <span
              className={`text-[11px] font-medium ${
                connected ? "text-[#4caf50]" : "text-[#f44336]"
              }`}
            >
              {connected ? "OBS 연결됨" : "OBS 미연결"}
            </span>
            <button
              onClick={onClose}
              className="bg-transparent border-none text-[#aaa] text-xl cursor-pointer leading-none hover:text-white transition-colors"
            >
              ✕
            </button>
          </div>
        </div>

        {/* tabs */}
        <div className="flex border-b border-white/10 px-5">
          {TABS.map((t) => (
            <button
              key={t.key}
              onClick={() => {
                setTab(t.key);
                if (t.key === "camera" || t.key === "setup") fetchDevices();
                if (t.key === "sources" || t.key === "camera" || t.key === "display") fetchSources();
              }}
              className={`px-4 py-2.5 bg-transparent border-none border-b-2 text-xs cursor-pointer transition-colors ${
                tab === t.key
                  ? "border-[#4a9eff] text-[#4a9eff] font-semibold"
                  : "border-transparent text-[#aaa] font-normal hover:text-white"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>

        {/* scene selector */}
        <div className="px-5 pt-3">
          <label className="text-[11px] text-[#888]">씬 선택</label>
          <select
            className={`${selectClass} mt-1`}
            value={selectedScene}
            onChange={(e) => setSelectedScene(e.target.value)}
          >
            <option value="">-- 씬 선택 --</option>
            {scenes.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>

        {/* body */}
        <div className="flex-1 overflow-y-auto px-5 py-4">
          {tab === "setup" ? (
            <div>
              {/* OBS 연결 설정 */}
              <div className="mb-5 pb-5 border-b border-white/10">
                <div className="text-[11px] text-[#888] mb-3 font-semibold uppercase tracking-wide">WebSocket 연결</div>
                <div className="flex gap-2 mb-2">
                  <div className="flex-1">
                    <label className="text-[11px] text-[#888]">서버 IP</label>
                    <input
                      className={`${selectClass} mt-1`}
                      value={obsIP}
                      onChange={(e) => setObsIP(e.target.value)}
                      placeholder="localhost"
                    />
                  </div>
                  <div className="w-24">
                    <label className="text-[11px] text-[#888]">포트</label>
                    <input
                      className={`${selectClass} mt-1`}
                      type="number"
                      value={obsPort}
                      onChange={(e) => setObsPort(parseInt(e.target.value) || 4455)}
                      placeholder="4455"
                    />
                  </div>
                </div>
                <div className="mb-3">
                  <label className="text-[11px] text-[#888]">비밀번호</label>
                  <input
                    className={`${selectClass} mt-1`}
                    type="password"
                    value={obsPassword}
                    onChange={(e) => setObsPassword(e.target.value)}
                    placeholder="WebSocket 비밀번호"
                  />
                </div>
                <button
                  onClick={handleConnect}
                  disabled={connectLoading}
                  className={`px-5 py-2 rounded-md text-white text-xs font-semibold transition-opacity border-none cursor-pointer ${
                    connected ? "bg-[#2e7d32] hover:bg-[#388e3c]" : "bg-[#204d87] hover:bg-[#2d5a8a]"
                  } ${connectLoading ? "opacity-60 cursor-default" : ""}`}
                >
                  {connectLoading ? "연결 중..." : connected ? "재연결" : "연결"}
                </button>
              </div>

              {!connected ? (
                <div className="text-[#666] text-xs text-center py-4">
                  연결 후 초기 설정을 진행할 수 있습니다.
                </div>
              ) : (
                <>
                  <p className="text-[#aaa] text-xs mb-4 leading-relaxed">
                    <span className="text-white font-semibold">camera</span> 씬과{" "}
                    <span className="text-white font-semibold">monitor</span> 씬이 없으면 자동 생성하고,
                    카메라·모니터 캡처·EP_Display 브라우저 소스를 자동으로 추가합니다.
                  </p>

                  {/* 카메라 디바이스 선택 */}
                  <div className="mb-4">
                    <div className="flex justify-between items-center mb-1">
                      <label className="text-[11px] text-[#888]">카메라 디바이스</label>
                      <button onClick={fetchDevices} className="bg-transparent border-none text-[#4a9eff] text-[11px] cursor-pointer">
                        {loadingDevices ? "조회 중..." : "새로고침"}
                      </button>
                    </div>
                    <select className={selectClass} value={selectedDevice} onChange={(e) => setSelectedDevice(e.target.value)}>
                      <option value="">-- 자동 감지 --</option>
                      {devices.map((d) => (
                        <option key={d.value} value={d.value}>{d.name}</option>
                      ))}
                    </select>
                    <p className="text-[#666] text-[10px] mt-1">선택하지 않으면 첫 번째 카메라를 자동으로 사용합니다.</p>
                  </div>

                  <div className="flex gap-2 items-center">
                    <button
                      onClick={handleSetupInitial}
                      disabled={setupLoading || setupDone}
                      className={`px-5 py-2 rounded-md text-white text-xs font-semibold border-none transition-opacity ${
                        setupDone
                          ? "bg-[#2e7d32] cursor-default opacity-80"
                          : setupLoading
                          ? "bg-[#204d87] opacity-60 cursor-default"
                          : "bg-[#204d87] cursor-pointer hover:bg-[#2d5a8a]"
                      }`}
                    >
                      {setupLoading ? "설정 중..." : setupDone ? "설정 완료 ✓" : "초기 설정 실행"}
                    </button>
                    {setupDone && (
                      <button
                        onClick={() => { setSetupDone(false); setSetupResult(null); }}
                        className="px-3 py-2 bg-white/10 border-none rounded-md text-[#aaa] text-xs cursor-pointer hover:bg-white/20 transition-colors"
                      >
                        재설정
                      </button>
                    )}
                  </div>
                </>
              )}
              {setupResult && (
                <div className="mt-4 space-y-2">
                  {!setupResult.success && (
                    <div className="p-3 bg-[rgba(244,67,54,0.1)] border border-[rgba(244,67,54,0.3)] rounded-md">
                      <div className="text-[#f44336] text-xs font-semibold mb-1">오류</div>
                      {(setupResult.warnings ?? []).map((w, i) => (
                        <div key={i} className="text-[#f44336] text-xs">• {w}</div>
                      ))}
                    </div>
                  )}
                  {(setupResult.scenes_created ?? []).length > 0 && (
                    <div className="p-3 bg-[rgba(76,175,80,0.1)] border border-[rgba(76,175,80,0.3)] rounded-md">
                      <div className="text-[#4caf50] text-xs font-semibold mb-1">생성된 씬</div>
                      {setupResult.scenes_created.map((s) => (
                        <div key={s} className="text-[#4caf50] text-xs">• {s}</div>
                      ))}
                    </div>
                  )}
                  {(setupResult.sources_created ?? []).length > 0 && (
                    <div className="p-3 bg-[rgba(74,158,255,0.08)] border border-[rgba(74,158,255,0.25)] rounded-md">
                      <div className="text-[#4a9eff] text-xs font-semibold mb-1">추가된 소스</div>
                      {setupResult.sources_created.map((s) => (
                        <div key={s} className="text-[#4a9eff] text-xs">• {s}</div>
                      ))}
                    </div>
                  )}
                  {setupResult.success && (setupResult.warnings ?? []).length > 0 && (
                    <div className="p-3 bg-[rgba(255,152,0,0.08)] border border-[rgba(255,152,0,0.25)] rounded-md">
                      <div className="text-[#ff9800] text-xs font-semibold mb-1">경고</div>
                      {setupResult.warnings.map((w, i) => (
                        <div key={i} className="text-[#ff9800] text-xs">• {w}</div>
                      ))}
                    </div>
                  )}
                  {setupResult.success &&
                    (setupResult.scenes_created ?? []).length === 0 &&
                    (setupResult.sources_created ?? []).length === 0 &&
                    (setupResult.warnings ?? []).length === 0 && (
                      <div className="text-[#888] text-xs">변경 사항 없음 (씬/소스가 이미 존재).</div>
                  )}
                </div>
              )}
            </div>
          ) : !connected ? (
            <div className="text-[#888] text-center py-10">
              OBS가 연결되지 않았습니다.
              <br />
              <span className="text-xs text-[#666]">설정 탭에서 WebSocket 연결 정보를 입력하세요.</span>
            </div>
          ) : tab === "logo" ? (
            /* ===== Logo Tab ===== */
            <div>
              {/* Upload zone */}
              <div
                onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
                onDragLeave={() => setDragOver(false)}
                onDrop={(e) => { e.preventDefault(); setDragOver(false); handleLogoUpload(e.dataTransfer.files); }}
                onClick={() => fileInputRef.current?.click()}
                className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer mb-4 transition-all ${
                  dragOver
                    ? "border-[#4a9eff] bg-[rgba(74,158,255,0.1)]"
                    : "border-white/20 bg-transparent hover:border-white/40"
                }`}
              >
                <div className="text-[#aaa] text-xs">
                  {logoUploaded
                    ? "로고 업로드 완료 (클릭하여 교체)"
                    : "로고 이미지를 드래그하거나 클릭하여 업로드"}
                </div>
                <div className="text-[#666] text-[11px] mt-1">PNG, JPG (최대 10MB)</div>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept="image/png,image/jpeg"
                  className="hidden"
                  onChange={(e) => handleLogoUpload(e.target.files)}
                />
              </div>

              {/* Position presets */}
              <div className="mb-4">
                <label className="text-[11px] text-[#888]">위치 프리셋</label>
                <div className="flex gap-2 mt-1.5">
                  {POSITIONS.map((p) => (
                    <button
                      key={p.key}
                      onClick={() => setLogoPosition(p.key)}
                      className={`flex-1 py-2 rounded-md text-xs cursor-pointer transition-all ${
                        logoPosition === p.key
                          ? "bg-[rgba(74,158,255,0.2)] border border-[#4a9eff] text-[#4a9eff]"
                          : "bg-white/[0.06] border border-white/15 text-[#aaa] hover:border-white/30"
                      }`}
                    >
                      {p.label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Scale slider */}
              <div className="mb-5">
                <label className="text-[11px] text-[#888]">
                  크기: {Math.round(logoScale * 100)}%
                </label>
                <input
                  type="range"
                  min={0.05}
                  max={0.5}
                  step={0.01}
                  value={logoScale}
                  onChange={(e) => setLogoScale(parseFloat(e.target.value))}
                  className="w-full mt-1.5"
                />
              </div>

              <div className="flex gap-2">
                <button onClick={handleLogoApply} disabled={busy} className={btnPrimaryClass}>
                  {busy ? "적용 중..." : "OBS에 적용"}
                </button>
                {logoApplied && logoItemId && (
                  <button
                    onClick={async () => {
                      const item = sources.find((s) => s.sourceName === "EP_Logo");
                      if (item) {
                        await apiClient.toggleOBSSource(selectedScene, item.sceneItemId, !item.enabled);
                        fetchSources();
                      }
                    }}
                    className="px-5 py-2 bg-white/10 border-none rounded-md text-white text-xs font-semibold cursor-pointer hover:bg-white/20 transition-colors"
                  >
                    {sources.find((s) => s.sourceName === "EP_Logo")?.enabled ? "숨기기" : "표시"}
                  </button>
                )}
              </div>
            </div>
          ) : tab === "display" ? (
            /* ===== Display Tab ===== */
            <div>
              <p className="text-[#aaa] text-xs mb-4 leading-relaxed">
                선택한 씬에 <span className="text-white font-semibold">EP_Display</span> 브라우저 소스를 자동으로 추가합니다.
                <br />기존 EP_Display가 있으면 삭제 후 재생성합니다.
              </p>

              <div className="mb-4">
                <label className="text-[11px] text-[#888]">Display URL</label>
                <input
                  className={`${selectClass} mt-1`}
                  value={displayURL}
                  onChange={(e) => setDisplayURL(e.target.value)}
                  placeholder="http://localhost:8080/display"
                />
                <p className="text-[#666] text-[10px] mt-1">
                  서버가 다른 포트나 IP에서 실행 중이면 URL을 변경하세요.
                </p>
              </div>

              <button onClick={handleSetupDisplay} disabled={busy} className={btnPrimaryClass}>
                {busy ? "설정 중..." : "Display 소스 설정"}
              </button>

              {displaySetupDone && (
                <div className="mt-4 p-3 bg-[rgba(76,175,80,0.1)] border border-[rgba(76,175,80,0.3)] rounded-md text-[#4caf50] text-xs">
                  EP_Display 소스가 "{selectedScene}" 씬에 추가되었습니다.
                </div>
              )}

              {sources.filter((s) => s.sourceName === "EP_Display").length > 0 && (
                <div className="mt-5">
                  <label className="text-[11px] text-[#888]">현재 Display 소스</label>
                  {sources
                    .filter((s) => s.sourceName === "EP_Display")
                    .map((item) => (
                      <SourceRow
                        key={item.sceneItemId}
                        item={item}
                        onToggle={() => handleToggle(item)}
                        onRemove={() => handleRemove(item)}
                      />
                    ))}
                </div>
              )}
            </div>
          ) : tab === "camera" ? (
            /* ===== Camera Tab ===== */
            <div>
              <div className="mb-3">
                <div className="flex justify-between items-center">
                  <label className="text-[11px] text-[#888]">카메라 디바이스</label>
                  <button
                    onClick={fetchDevices}
                    className="bg-transparent border-none text-[#4a9eff] text-[11px] cursor-pointer"
                  >
                    {loadingDevices ? "조회 중..." : "새로고침"}
                  </button>
                </div>
                <select
                  className={`${selectClass} mt-1`}
                  value={selectedDevice}
                  onChange={(e) => setSelectedDevice(e.target.value)}
                >
                  <option value="">
                    {loadingDevices ? "조회 중..." : "-- 디바이스 선택 --"}
                  </option>
                  {devices.map((d) => (
                    <option key={d.value} value={d.value}>{d.name}</option>
                  ))}
                </select>
              </div>

              <div className="mb-4">
                <label className="text-[11px] text-[#888]">소스 이름</label>
                <input
                  className={`${selectClass} mt-1`}
                  value={cameraName}
                  onChange={(e) => setCameraName(e.target.value)}
                  placeholder="카메라"
                />
              </div>

              <button onClick={handleAddCamera} disabled={busy} className={btnPrimaryClass}>
                {busy ? "추가 중..." : "추가"}
              </button>

              {sources.filter(
                (s) => s.inputKind === "av_capture_input_v2" || s.inputKind === "dshow_input"
              ).length > 0 && (
                <div className="mt-5">
                  <label className="text-[11px] text-[#888]">현재 카메라 소스</label>
                  {sources
                    .filter(
                      (s) => s.inputKind === "av_capture_input_v2" || s.inputKind === "dshow_input"
                    )
                    .map((item) => (
                      <SourceRow
                        key={item.sceneItemId}
                        item={item}
                        onToggle={() => handleToggle(item)}
                        onRemove={() => handleRemove(item)}
                      />
                    ))}
                </div>
              )}
            </div>
          ) : (
            /* ===== All Sources Tab ===== */
            <div>
              {loadingSources ? (
                <div className="text-[#888] text-center py-5">로딩 중...</div>
              ) : sources.length === 0 ? (
                <div className="text-[#666] text-center py-5">
                  {selectedScene ? "소스가 없습니다." : "씬을 선택하세요."}
                </div>
              ) : (
                sources.map((item) => (
                  <SourceRow
                    key={item.sceneItemId}
                    item={item}
                    onToggle={() => handleToggle(item)}
                    onRemove={() => handleRemove(item)}
                  />
                ))
              )}
            </div>
          )}
        </div>
      </div>

      {/* toast */}
      {toastMsg && (
        <div
          className={`fixed top-6 left-1/2 -translate-x-1/2 px-6 py-2.5 rounded-lg text-white text-xs z-[11100] shadow-lg ${
            toastMsg.type === "error" ? "bg-[#dc3545]" : "bg-[#204d87]"
          }`}
        >
          {toastMsg.msg}
        </div>
      )}
    </div>
  );
}

function SourceRow({
  item,
  onToggle,
  onRemove,
}: {
  item: OBSSourceItem;
  onToggle: () => void;
  onRemove: () => void;
}) {
  return (
    <div className="flex items-center justify-between px-2.5 py-2 mt-1.5 bg-white/[0.04] rounded-md border border-white/[0.08]">
      <div className="flex-1 min-w-0">
        <div
          className={`text-xs overflow-hidden text-ellipsis whitespace-nowrap ${
            item.enabled ? "text-white" : "text-[#666]"
          }`}
        >
          {item.sourceName}
        </div>
        <div className="text-[#666] text-[10px] mt-0.5">{item.inputKind}</div>
      </div>
      <div className="flex gap-1.5 flex-shrink-0">
        <button
          onClick={onToggle}
          className={`px-2.5 py-1 rounded text-[11px] cursor-pointer transition-colors ${
            item.enabled
              ? "bg-[rgba(76,175,80,0.15)] border border-[rgba(76,175,80,0.3)] text-[#4caf50]"
              : "bg-white/[0.06] border border-white/10 text-[#888]"
          }`}
        >
          {item.enabled ? "ON" : "OFF"}
        </button>
        <button
          onClick={onRemove}
          className="px-2.5 py-1 rounded text-[11px] cursor-pointer bg-[rgba(255,60,60,0.1)] border border-[rgba(255,60,60,0.2)] text-[#ff6b6b] hover:bg-[rgba(255,60,60,0.2)] transition-colors"
        >
          삭제
        </button>
      </div>
    </div>
  );
}
