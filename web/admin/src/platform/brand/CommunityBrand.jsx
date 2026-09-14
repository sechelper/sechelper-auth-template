import { adminAppConfig } from "../config/runtime.js";

export function CommunityBrand({ href = "/" }) {
  const systemName = adminAppConfig.systemName.trim();
  const primaryName = `助安社区 - ${systemName}`;

  return (
    <a className="sechelper-brand" href={href} aria-label={`${primaryName} · SECHELPER COMMUNITY`}>
      <img className="sechelper-brand__logo" src="/admin/assets/logo/logo.png" alt="SECHELPER COMMUNITY" />
      <span className="sechelper-brand__copy">
        <span className="sechelper-brand__cn" title={primaryName}>{primaryName}</span>
        <span className="sechelper-brand__en">SECHELPER COMMUNITY</span>
      </span>
    </a>
  );
}
