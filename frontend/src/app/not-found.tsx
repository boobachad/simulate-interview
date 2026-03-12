import Link from "next/link";
import { Home, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function Custom404() {
  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-muted">
            <span className="text-4xl font-bold text-muted-foreground">
              404
            </span>
          </div>
          <CardTitle>Page Not Found</CardTitle>
          <CardDescription>
            The page you&apos;re looking for doesn&apos;t exist or has been
            moved.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          <Button asChild variant="default" className="w-full">
            <Link href="/start">
              <Home className="mr-2 h-4 w-4" />
              Go to Home
            </Link>
          </Button>
          <Button asChild variant="outline" className="w-full">
            <button type="button" onClick={() => window.history.back()}>
              <ArrowLeft className="mr-2 h-4 w-4" />
              Go Back
            </button>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
