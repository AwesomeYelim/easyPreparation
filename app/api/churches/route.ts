// app/api/churches/route.ts
import { NextResponse } from "next/server";
import pool from "@/lib/db";

type Church = {
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
    const { rows }: { rows: Church[] } = await pool.query(
      "SELECT * FROM churches WHERE email = $1 LIMIT 1",
      [email]
    );

    if (rows.length === 0) {
      return NextResponse.json({ error: "Church not found" }, { status: 404 });
    }

    return NextResponse.json(rows[0]);
  } catch (error) {
    console.error(error);
    return NextResponse.json(
      { error: "Failed to fetch data" },
      { status: 500 }
    );
  }
}

export async function POST(request: Request) {
  // 요청 본문에서 title과 content 값을 추출
  const { title, content }: { title: string; content: string } =
    await request.json();

  // title과 content가 없는 경우 에러 반환
  if (!title || !content) {
    return NextResponse.json(
      { error: "Title and content are required" },
      { status: 400 }
    );
  }

  try {
    const result = await pool.query(
      "INSERT INTO posts (title, content) VALUES ($1, $2) RETURNING *",
      [title, content]
    );
    return NextResponse.json(result.rows[0], { status: 201 }); // 삽입된 데이터 반환
  } catch (error) {
    console.error(error);
    return NextResponse.json(
      { error: "Failed to insert data" },
      { status: 500 }
    );
  }
}
