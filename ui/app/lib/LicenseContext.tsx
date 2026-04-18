'use client';
import { createContext, useContext, useCallback, useEffect, useState, ReactNode } from 'react';
import { useSetRecoilState } from 'recoil';
import { licenseState } from '../recoilState';
import { apiClient } from './apiClient';
import { LicenseStatus, LicenseFeature } from '../types';

interface LicenseContextType {
  license: LicenseStatus;
  hasFeature: (feature: LicenseFeature) => boolean;
  refresh: () => Promise<void>;
  isLoading: boolean;
}

const DEFAULT_LICENSE: LicenseStatus = {
  plan: 'free',
  features: [],
  expires_at: null,
  days_remaining: 0,
  device_id: '',
  grace_period: false,
  is_active: false,
};

const LicenseContext = createContext<LicenseContextType>({
  license: DEFAULT_LICENSE,
  hasFeature: () => false,
  refresh: async () => {},
  isLoading: true,
});

export function useLicense() {
  return useContext(LicenseContext);
}

export function LicenseProvider({ children }: { children: ReactNode }) {
  const [license, setLicense] = useState<LicenseStatus>(DEFAULT_LICENSE);
  const [isLoading, setIsLoading] = useState(true);
  const setRecoilLicense = useSetRecoilState(licenseState);

  const refresh = useCallback(async () => {
    try {
      const status = await apiClient.getLicenseStatus();
      setLicense(status);
      setRecoilLicense(status);
    } catch (err) {
      console.error('License status fetch failed:', err);
    } finally {
      setIsLoading(false);
    }
  }, [setRecoilLicense]);

  useEffect(() => {
    refresh();
    // 1시간마다 자동 갱신
    const interval = setInterval(refresh, 60 * 60 * 1000);
    return () => clearInterval(interval);
  }, [refresh]);

  const hasFeature = useCallback(
    (feature: LicenseFeature): boolean => {
      if (process.env.NEXT_PUBLIC_DEV_MODE === 'true') return true;
      return license.features.includes(feature);
    },
    [license.features],
  );

  return (
    <LicenseContext.Provider value={{ license, hasFeature, refresh, isLoading }}>
      {children}
    </LicenseContext.Provider>
  );
}
