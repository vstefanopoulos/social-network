"use client";

import { LiveSocketProvider } from "@/context/LiveSocketContext";

export default function LiveSocketWrapper({ children, wsUrl }) {
    return <LiveSocketProvider wsUrl={wsUrl}>{children}</LiveSocketProvider>;
}
