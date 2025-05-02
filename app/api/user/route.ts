// app/api/churches/route.ts
import { NextResponse } from "next/server";
import pool from "@/lib/db";

type User = {
  id: number;
  name: string;
  email: string;
};

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const email = searchParams.get("email");

  if (!email) {
    return NextResponse.json({ error: "Email is required" }, { status: 400 });
  }

  try {
    const result = await pool.query(
      `
      SELECT 
        c.id,
        c.name,
        c.english_name,
        c.email,
        l.license_key AS figma_key,
        l.license_token AS figma_token
      FROM churches c
      LEFT JOIN licenses l ON c.id = l.church_id
      WHERE c.email = $1
      ORDER BY l.issued_at DESC
      LIMIT 1
      `,
      [email]
    );

    if (result.rows.length === 0) {
      return NextResponse.json(
        { error: "User church info not found" },
        { status: 404 }
      );
    }

    const row = result.rows[0];

    const response = {
      id: row.id,
      name: row.name,
      english_name: row.english_name,
      email: row.email,
      figmaInfo: {
        key: row.figma_key || "",
        token: row.figma_token || "",
      },
    };

    return NextResponse.json(response);
  } catch (error) {
    console.error("Database query failed:", error);
    return NextResponse.json(
      { error: "Failed to fetch data" },
      { status: 500 }
    );
  }
}
export async function POST(req: Request) {
  const { name, english_name, email } = await req.json();

  if (!name || !english_name || !email) {
    return NextResponse.json(
      { error: "모든 필드를 입력해주세요." },
      { status: 400 }
    );
  }

  try {
    const existing = await pool.query(
      "SELECT 1 FROM churches WHERE email = $1 LIMIT 1",
      [email]
    );

    if (existing.rows.length > 0) {
      // 이미 존재 → UPDATE
      await pool.query(
        "UPDATE churches SET name = $1, english_name = $2 WHERE email = $3",
        [name, english_name, email]
      );
    } else {
      // 존재하지 않음 → INSERT
      await pool.query(
        "INSERT INTO churches (name, english_name, email) VALUES ($1, $2, $3)",
        [name, english_name, email]
      );
    }

    return NextResponse.json({ success: true });
  } catch (err) {
    console.error("DB 오류:", err);
    return NextResponse.json({ error: "DB 작업 실패" }, { status: 500 });
  }
}
