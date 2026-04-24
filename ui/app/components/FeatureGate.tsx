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
      {fallback !== undefined ? fallback : (
        <span className="inline-flex items-center gap-1 px-2.5 py-1 bg-blue-50 text-blue-600 text-xs font-bold rounded-full border border-blue-200">
          <span className="material-symbols-outlined" style={{ fontSize: '12px' }}>lock</span>
          Pro only
        </span>
      )}
    </>
  );
}

export function useFeature(feature: LicenseFeature): boolean {
  const { hasFeature } = useLicense();
  return hasFeature(feature);
}
