import Navbar from "@/components/layout/Navbar";
import LiveSocketWrapper from "@/components/providers/LiveSocketWrapper";
import { ToastProvider } from "@/context/ToastContext";

export const dynamic = 'force-dynamic';

export default function MainLayout({ children }) {
    const wsUrl = process.env.LIVE;

    return (
        <LiveSocketWrapper wsUrl={wsUrl}>
            <ToastProvider>
                <div className="min-h-screen flex flex-col bg-(--muted)/6">
                    <Navbar />
                    <main className="flex-1 w-full">
                        {children}
                    </main>
                </div>
            </ToastProvider>
        </LiveSocketWrapper>
    );
}