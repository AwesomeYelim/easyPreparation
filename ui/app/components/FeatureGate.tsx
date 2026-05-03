'use client';
import { ReactNode, useState, useEffect } from 'react';
import { useLicense } from '@/lib/LicenseContext';
import { LicenseFeature } from '@/types';

interface FeatureGateProps {
  feature: LicenseFeature;
  children: ReactNode;
  fallback?: ReactNode;
}

function ProBadge() {
  const { openLicensePanel } = useLicense();
  return (
    <button
      type="button"
      onClick={openLicensePanel}
      title="Pro 플랜으로 업그레이드하면 이 기능을 사용할 수 있습니다"
      className="inline-flex items-center gap-1 px-2.5 py-1 bg-[#1a3a5e] text-[#60a5fa] text-xs font-bold rounded-full border border-[#1e5a8a] hover:bg-[#1e4a7a] hover:border-[#3b82f6] transition-colors cursor-pointer"
    >
      <span className="material-symbols-outlined" style={{ fontSize: '12px' }}>lock</span>
      Pro 전용
    </button>
  );
}

export default function FeatureGate({ feature, children, fallback }: FeatureGateProps) {
  const { hasFeature } = useLicense();
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);

  // SSR + initial hydration: always render fallback to guarantee server/client match
  if (!mounted) return <>{fallback !== undefined ? fallback : <ProBadge />}</>;

  if (hasFeature(feature)) return <>{children}</>;
  return <>{fallback !== undefined ? fallback : <ProBadge />}</>;
}

export function useFeature(feature: LicenseFeature): boolean {
  const { hasFeature } = useLicense();
  return hasFeature(feature);
}
