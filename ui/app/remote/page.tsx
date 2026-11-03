"use client";
import { useEffect } from "react";
import { useRecoilState } from "recoil";
import ProSequencePanel from "@/components/ProSequencePanel";
import InspectorPreviewTab from "@/components/inspector/InspectorPreviewTab";
import { bulletinPreviewState } from "@/recoilState";
import { usePageVisible } from "@/lib/usePageVisible";

/**
 * 분리된 순서 패널 창 (/remote) — 데스크톱 앱이 네이티브 창으로 띄운다(표시·숨김·닫힘은 Go 가 처리).
 * 위: Display 미리보기, 아래: 예배 순서.
 */
export default function RemotePanelPage() {
  const [bulletinPreview, setBulletinPreview] = useRecoilState(bulletinPreviewState);
  // 미리 띄워 둔 숨은 창에서는 Display 미리보기 iframe 을 내려 둔다 (보일 때만 로드)
  const visible = usePageVisible();

  useEffect(() => {
    document.title = "easyPreparation — 예배 순서";
  }, []);

  return (
    <div className="bg-pro-bg text-pro-text" style={{ height: "100vh", display: "flex", flexDirection: "column" }}>
      {/* 위: Display 실시간 미리보기 (메인 창 Studio 의 미리보기 탭과 같은 컴포넌트) */}
      <div className="flex-shrink-0 p-2 border-b border-pro-border bg-pro-surface" data-testid="remote-preview">
        {visible && (
          <InspectorPreviewTab bulletinPreview={bulletinPreview} setBulletinPreview={setBulletinPreview} />
        )}
      </div>
      {/* 아래: 예배 순서 */}
      <ProSequencePanel standalone />
    </div>
  );
}
