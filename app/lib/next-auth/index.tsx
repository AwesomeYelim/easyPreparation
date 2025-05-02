"use client";

import React from "react";
import { SessionProvider } from "next-auth/react";
import CompleteProfile from "@/lib/next-auth/page";

interface Props {
  children: React.ReactNode;
}

const AuthProvider = ({ children }: Props) => {
  return (
    <SessionProvider>
      <CompleteProfile />
      {children}
    </SessionProvider>
  );
};

export default AuthProvider;
