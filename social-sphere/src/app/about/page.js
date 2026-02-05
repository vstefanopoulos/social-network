import Link from "next/link";

export default function About() {
    return (
        <div className="page-container ">
            {/* Navigation */}
            <nav className="section-border">
                <div className="max-w-7xl mx-auto px-6 py-6">
                    <div className="flex items-center justify-between">
                        <Link href="/" className="flex items-center gap-3">
                            <span className="text-base font-medium tracking-tight">SocialSphere</span>
                        </Link>
                        <div className="flex items-center gap-8">
                            <Link
                                href="/login"
                                className="text-sm link-muted"
                                prefetch={false}
                            >
                                Sign In
                            </Link>
                            <Link
                                href="/register"
                                className="link-primary text-sm"
                            >
                                Get Started
                            </Link>
                        </div>
                    </div>
                </div>
            </nav>

            <section className="section-border bg-(--muted)/10">
                <div className="flex items-center justify-center h-screen">
                    <h1 className="text-5xl">Page under construction</h1>
                </div>
            </section>
        </div>
    );
}