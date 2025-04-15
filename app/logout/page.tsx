"use client";

import { signOut } from "next-auth/react";

const Logout = () => {
  const handleLogout = async () => {
    await signOut({ callbackUrl: "/login" }); // 로그아웃 후 로그인 페이지로 리디렉션
  };

  return (
    <div>
      <button onClick={handleLogout}>Logout</button>
    </div>
  );
};

export default Logout;
