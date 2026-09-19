export function CommunityBrand({ href = "/" }) {
  const brandName = "助安社区";

  return (
    <a className="sechelper-brand" href={href} aria-label={brandName}>
      <img className="sechelper-brand__logo" src="/admin/assets/logo/logo.png" alt={brandName} />
      <span className="sechelper-brand__copy">
        <span className="sechelper-brand__cn" title={brandName}>{brandName}</span>
        <span className="sechelper-brand__en">SECHELPER COMMUNITY</span>
      </span>
    </a>
  );
}
