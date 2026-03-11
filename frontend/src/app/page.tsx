'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import { Navbar } from '@/components/Navbar';

export default function Home() {
  const router = useRouter();

  return (
    <div className="min-h-screen bg-bg-primary flex flex-col">
      <Navbar breadcrumbItems={[{ label: 'home' }]} />

      <div className="flex-1 flex items-center justify-center">

        <div className="text-center max-w-2xl px-4">
          <h1 className="text-6xl font-bold text-content-primary mb-6">
            Interview Platform
          </h1>
          <p className="text-xl text-content-secondary mb-12">
            Head to Interview like experience. select your focus area and start practicing.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <button
              onClick={() => router.push('/onboarding')}
              className="py-4 px-8 bg-primary text-primary-foreground text-lg font-semibold rounded-lg hover:bg-primary/90 transition-all shadow-lg hover:shadow-xl"
            >
              Get Started
            </button>
            <button
              onClick={() => router.push('/start')}
              className="py-4 px-8 bg-surface-elevated text-content-primary text-lg font-semibold rounded-lg hover:bg-surface-hover transition-all border border-border-primary"
            >
              Start Practice
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
