import { NextResponse } from "next/server";
import pool from "@/lib/db"; // pool은 PostgreSQL 연결을 담당하는 객체입니다.

type Post = {
  id: number;
  title: string;
  content: string;
};

export async function GET() {
  try {
    const { rows }: { rows: Post[] } = await pool.query("SELECT * FROM posts");
    return NextResponse.json(rows); // 데이터를 JSON 형식으로 반환
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
