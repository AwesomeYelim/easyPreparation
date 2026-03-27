import { NextRequest, NextResponse } from "next/server";
import { writeFile, mkdir } from "fs/promises";
import path from "path";

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const { target, targetInfo } = body;

    if (!target || !targetInfo) {
      return NextResponse.json({ error: "Missing data" }, { status: 400 });
    }

    // 파일 경로 (예: app/data/main_worship.json)
    const dirPath = path.join(process.cwd(), "app", "data");
    const filePath = path.join(dirPath, `${target}.json`);

    // 디렉토리 없으면 생성
    await mkdir(dirPath, { recursive: true });

    // 파일 저장
    await writeFile(filePath, JSON.stringify(targetInfo, null, 2), "utf-8");

    return NextResponse.json({ message: "File saved", path: filePath });
  } catch (err: any) {
    console.error("API Error:", err);
    return NextResponse.json({ error: "Failed to save file" }, { status: 500 });
  }
}
