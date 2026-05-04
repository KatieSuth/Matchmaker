"use client";

import { useEffect } from "react";
import { capturePostLoginRedirectFromWindowSearch } from "@/app/_lib/postLoginRedirect";

/** Runs on the landing page so ?next= from middleware is stored before Discord OAuth. */
export function PostLoginRedirectCapture() {
  useEffect(() => {
    capturePostLoginRedirectFromWindowSearch();
  }, []);
  return null;
}
