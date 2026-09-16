export const GITHUB = "https://github.com/AwesomeYelim/easyPreparation";
export const RELEASES = `${GITHUB}/releases`;
export const RELEASE_LATEST = `${RELEASES}/latest`;
export const RELEASE_BASE = `${RELEASE_LATEST}/download`;
export const ISSUES = `${GITHUB}/issues`;
export const README = `${GITHUB}#readme`;
export const USAGE_GUIDE = `${GITHUB}/blob/master/docs/USAGE_GUIDE.md`;

export const INSTALL_MAC = `curl -fsSL ${RELEASE_BASE}/install-mac.sh | bash`;
export const INSTALL_WIN = `irm ${RELEASE_BASE}/install-windows.ps1 | iex`;

// Asset names must match .github/workflows/release.yml output
export const ASSETS = {
  desktopMac: "easyPreparation_desktop_darwin_arm64.dmg",
  desktopWin: "easyPreparation_desktop_windows_amd64_setup.exe",
  desktopLinux: "easyPreparation_desktop_linux_amd64",
  serverMacArm: "easyPreparation_server_darwin_arm64",
  serverMacIntel: "easyPreparation_server_darwin_amd64",
  serverLinux: "easyPreparation_server_linux_amd64",
  serverWin: "easyPreparation_server_windows_amd64.exe",
  checksums: "checksums.txt",
} as const;

export const asset = (name: string) => `${RELEASE_BASE}/${name}`;
