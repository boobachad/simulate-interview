'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Navbar } from '@/components/Navbar';

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      router.push('/start');
    }
  }, [router]);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Navbar breadcrumbItems={[{ label: 'home' }]} showAuth={false} />

      <div className="flex-1 flex items-center justify-center">
        <div className="text-center max-w-2xl px-4">
          <h1 className="text-6xl font-bold mb-6">
            Interview Platform
          </h1>
          <p className="text-xl text-muted-foreground mb-12">
            Practice coding interviews with personalized problems based on your LeetCode and Codeforces stats
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <button
              onClick={() => router.push('/login')}
              className="py-4 px-8 bg-primary text-primary-foreground text-lg font-semibold rounded-lg hover:bg-primary/90 transition-all shadow-lg hover:shadow-xl"
            >
              Get Started
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
