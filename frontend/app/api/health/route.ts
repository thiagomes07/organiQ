import { NextResponse } from "next/server";

/**
 * Health Check Endpoint
 *
 * Retorna o status da aplicação para monitoramento
 * Útil para load balancers, Kubernetes, etc.
 */
export async function GET() {
  try {
    return NextResponse.json(
      {
        status: "healthy",
        timestamp: new Date().toISOString(),
        service: "organiQ Frontend",
        version: process.env.NEXT_PUBLIC_APP_VERSION || "1.0.0",
        environment: process.env.NODE_ENV,
      },
      { status: 200 }
    );
  } catch (error) {
    return NextResponse.json(
      {
        status: "unhealthy",
        timestamp: new Date().toISOString(),
        error: error instanceof Error ? error.message : "Unknown error",
      },
      { status: 503 }
    );
  }
}

// Permite apenas GET
export async function POST() {
  return NextResponse.json({ error: "Method not allowed" }, { status: 405 });
}
