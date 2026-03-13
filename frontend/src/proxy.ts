
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const publicPaths = ['/', '/login'];

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  
  // Check if path starts with any public path
  if (publicPaths.some(path => pathname === path || pathname.startsWith(path + '/'))) {
    return NextResponse.next();
  }

  const token = request.cookies.get('auth_token')?.value;
  
  if (!token) {
    const loginUrl = new URL('/login', request.url);
    // Preserve query string in redirect
    loginUrl.searchParams.set('redirect', pathname + request.nextUrl.search);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
