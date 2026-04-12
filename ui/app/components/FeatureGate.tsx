'use client';
import { ReactNode } from 'react';
import { useLicense } from '@/lib/LicenseContext';
import { LicenseFeature } from '@/types';

interface FeatureGateProps {
  feature: LicenseFeature;
  children: ReactNode;
  fallback?: ReactNode;
}

export default function FeatureGate({ feature, children, fallback }: FeatureGateProps) {
  const { hasFeature } = useLicense();
  if (hasFeature(feature)) return <>{children}</>;
  return (
    <>
      {fallback ?? (
        <div style={{ textAlign: 'center', padding: '24px', color: '#888' }}>
          <div style={{ fontSize: '24px', marginBottom: '8px' }}>Pro</div>
          <p>이 기능은 Pro 플랜에서 사용할 수 있습니다.</p>
        </div>
      )}
    </>
  );
}

export function useFeature(feature: LicenseFeature): boolean {
  const { hasFeature } = useLicense();
  return hasFeature(feature);
}
