"use client";

import Link from "next/link";
import classNames from "classnames";
import { usePathname } from "next/navigation";

export default function NavLink({
  href,
  children,
}: {
  href: string;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const isActive = pathname?.includes(href);

  return (
    <Link href={href} className={classNames({ active: isActive })}>
      {children}
    </Link>
  );
}
