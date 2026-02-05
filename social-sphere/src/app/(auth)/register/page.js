import Link from "next/link";
import RegisterForm from "@/components/forms/RegisterForm";

export default function RegisterPage() {
    return (
        <div className="page-container min-h-screen flex items-center justify-center px-6 py-12">
            <div className="w-full max-w-5xl">
                {/* Header */}
                <div className="mb-10 text-center">
                    <h1 className="heading-md mb-2">
                        Join SocialSphere
                    </h1>
                    <p className="text-muted text-sm">
                        Create your account and start connecting
                    </p>
                </div>

                {/* Form - Split into two columns */}
                <RegisterForm />

                {/* Footer */}
                <p className="mt-2 text-sm text-center text-muted">
                    Already have an account?{" "}
                    <Link href="/login" className="link-primary">
                        Sign in
                    </Link>
                </p>
            </div>
        </div>
    );
}
