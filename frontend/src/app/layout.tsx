import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";

const geistSans = localFont({
  src: "./fonts/GeistVF.woff",
  variable: "--font-geist-sans",
});
const geistMono = localFont({
  src: "./fonts/GeistMonoVF.woff",
  variable: "--font-geist-mono",
});

export const metadata: Metadata = {
  title: "Interview Platform - C++ Practice",
  description: "Competitive programming interview practice platform",
};

import { Toaster } from "@/components/ui/sonner";
import { ThemeProvider } from "@/components/Themeprovider";
import { AppFooter } from "@/components/layout/app-footer";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} font-sans antialiased`}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="dark"
          enableSystem
          disableTransitionOnChange
        >
          <div className="flex flex-col min-h-screen">
            <main className="flex-1">
              {children}
            </main>
            <AppFooter />
          </div>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
