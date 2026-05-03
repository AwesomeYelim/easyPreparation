'use client';
import LicensePanel from './LicensePanel';
import { useLicense } from '@/lib/LicenseContext';

export default function LicensePanelMount() {
  const { isLicensePanelOpen, closeLicensePanel } = useLicense();
  return <LicensePanel open={isLicensePanelOpen} onClose={closeLicensePanel} />;
}
