import Link from "next/link";
import LoginForm from "@/components/forms/LoginForm";

export default function LoginPage() {
    return (
        <div className="page-container min-h-screen flex items-center justify-center px-6 py-12">
            <div className="w-full max-w-md">
                {/* Header */}
                <div className="mb-12">
                    <h1 className="text-center heading-md mb-3">
                        Welcome back
                    </h1>
                    <p className="text-center text-neutral-600">
                        Sign in to continue to SocialSphere
                    </p>
                </div>

                {/* Form */}
                <LoginForm />

                {/* Footer */}
                <p className="mt-2 text-sm text-neutral-600 text-center">
                    Don't have an account?{" "}
                    <Link href="/register" className="link-primary">
                        Sign up
                    </Link>
                </p>
            </div>
        </div>
    );
}
