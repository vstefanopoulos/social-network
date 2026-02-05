import { NextResponse } from "next/server";

// Routes that require authentication (protected routes)
const protectedRoutes = ["/feed", "/groups", "/posts", "/profile"];

// Routes that should NOT be accessible when logged in
const authRoutes = ["/login", "/register"];

export function proxy(request) {
  const { pathname } = request.nextUrl;

  // Check if user has JWT cookie (is logged in)
  const token = request.cookies.get("jwt")?.value;
  const isAuthenticated = !!token;

  // Check if current path is a protected route
  const isProtectedRoute = protectedRoutes.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  // Check if current path is an auth route (login/register)
  const isAuthRoute = authRoutes.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  // If user is NOT authenticated and trying to access protected route
  // → Redirect to login
  if (!isAuthenticated && isProtectedRoute) {
    const loginUrl = new URL("/login", request.url);
    // Store the attempted URL to redirect back after login
    loginUrl.searchParams.set("callbackUrl", pathname);
    return NextResponse.redirect(loginUrl);
  }

  // If user IS authenticated and trying to access auth routes (login/register)
  // → Redirect to feed
  if (isAuthenticated && isAuthRoute) {
    return NextResponse.redirect(new URL("/feed/public", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - api routes (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico, images, etc.
     */
    "/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
  ],
};
