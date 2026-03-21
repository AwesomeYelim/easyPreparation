"""
문서(docx) 생성 스크립트 템플릿
usage: python generate_docx.py <출력경로> [--data <json_file>]
"""

import sys
import os
import io
import argparse
import json
import tempfile

import matplotlib
matplotlib.use("Agg")
try:
    import matplotlib.pyplot as plt
    matplotlib.rcParams["font.family"] = "Malgun Gothic"
    matplotlib.rcParams["axes.unicode_minus"] = False
except Exception:
    pass

from docx import Document
from docx.shared import Pt, Inches, Cm, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.oxml.ns import qn
from docx.oxml import OxmlElement

# 공식 워드 템플릿 경로 (tools/templates/doc_template.docx)
_TEMPLATE = os.path.join(os.path.dirname(__file__), "..", "templates", "doc_template.docx")


def _load_document():
    """템플릿 기반 Document 반환. 기존 내용 초기화."""
    if os.path.exists(_TEMPLATE):
        doc = Document(_TEMPLATE)
        # sectPr(섹션/페이지 설정)는 유지하고 나머지 body 내용만 제거
        body = doc.element.body
        sectPr = body.find(qn("w:sectPr"))
        for child in list(body):
            if child is not sectPr:
                body.remove(child)
        return doc
    return Document()


def set_cell_bg(cell, hex_color: str):
    tc = cell._tc
    tcPr = tc.get_or_add_tcPr()
    shd = OxmlElement("w:shd")
    shd.set(qn("w:val"), "clear")
    shd.set(qn("w:color"), "auto")
    shd.set(qn("w:fill"), hex_color)
    tcPr.append(shd)


def styled_table(doc, headers, rows, header_hex="1A335C", stripe_hex="EBF5FB"):
    table = doc.add_table(rows=1, cols=len(headers))
    # 템플릿에 "Table Grid" 없으면 "Table Normal" 폴백
    try:
        table.style = "Table Grid"
    except KeyError:
        try:
            table.style = "Table Normal"
        except KeyError:
            pass
    hdr = table.rows[0].cells
    for i, h in enumerate(headers):
        hdr[i].text = h
        set_cell_bg(hdr[i], header_hex)
        p = hdr[i].paragraphs[0]
        p.alignment = WD_ALIGN_PARAGRAPH.CENTER
        run = p.runs[0] if p.runs else p.add_run(h)
        run.font.bold = True
        run.font.color.rgb = RGBColor(0xFF, 0xFF, 0xFF)
        run.font.size = Pt(9)
    for ri, row_data in enumerate(rows):
        row = table.add_row()
        bg = stripe_hex if ri % 2 == 0 else "FFFFFF"
        for ci, val in enumerate(row_data):
            row.cells[ci].text = str(val)
            set_cell_bg(row.cells[ci], bg)
            row.cells[ci].paragraphs[0].runs[0].font.size = Pt(8.5)
    return table


def build(output_path: str, data: dict):
    doc = _load_document()

    # 템플릿 페이지 설정 유지 (템플릿 없는 폴백 환경용 기본값)
    section = doc.sections[0]
    if section.page_width.cm < 1:
        section.page_width    = Cm(21)
        section.page_height   = Cm(29.7)
        section.left_margin   = Cm(2.5)
        section.right_margin  = Cm(2.5)
        section.top_margin    = Cm(2.5)
        section.bottom_margin = Cm(2.5)

    # ── 여기서부터 문서 내용 추가 ──────────────────────────────
    def h1(text):
        p = doc.add_heading(text, level=1)
        p.runs[0].font.color.rgb = RGBColor(0x1A, 0x23, 0x7E)

    def h2(text):
        p = doc.add_heading(text, level=2)
        p.runs[0].font.color.rgb = RGBColor(0x1B, 0x5E, 0x20)

    def body(text):
        p = doc.add_paragraph(text)
        p.runs[0].font.size = Pt(10)

    # 타이틀
    title = doc.add_heading("GSAC 프로젝트 개요", level=0)
    title.alignment = WD_ALIGN_PARAGRAPH.CENTER
    sub = doc.add_paragraph("Go Security Agent Cluster  |  2026-03-05")
    sub.alignment = WD_ALIGN_PARAGRAPH.CENTER
    sub.runs[0].font.color.rgb = RGBColor(0x78, 0x90, 0x9C)
    doc.add_paragraph()

    # 1. 프로젝트 개요
    h1("1. 프로젝트 개요")
    body(
        "GSAC(Go Security Agent Cluster)는 Go 언어 기반의 보안 에이전트 시스템입니다. "
        "SA(System Agent)를 중심으로 여러 보안 모듈을 오케스트레이션하며, "
        "AM(Agent Manager) 서버와 HTTP/gRPC로 통신하여 보안 수집·분석 작업을 수행합니다."
    )
    doc.add_paragraph()

    # 2. 코드베이스 통계
    h1("2. 코드베이스 통계")
    styled_table(doc,
        ["항목", "수치"],
        [
            ("총 모듈",  "24"),
            ("총 파일",  "588"),
            ("총 함수",  "2,378"),
            ("총 라인",  "48,821"),
            ("예상 토큰", "195,284"),
        ],
        header_hex="1A237E",
    )
    doc.add_paragraph()

    # 3. 주요 모듈
    h1("3. 주요 모듈")
    styled_table(doc,
        ["모듈", "경로", "파일", "함수", "라인", "역할"],
        [
            ("SA",  "cmd/sa/",           "33",  "187",  "4,438",  "시스템 에이전트 오케스트레이터"),
            ("AD",  "cmd/modules/ad/",   "82",  "407",  "4,960",  "Active Directory 수집"),
            ("AW",  "cmd/modules/aw/",   "58",  "314",  "7,876",  "Anti-Web 보안 분석"),
            ("AA",  "cmd/modules/aa/",   "23",   "92",  "2,344",  "Asset 정보 수집"),
            ("AM",  "cmd/am/",           "29",  "115",  "2,033",  "에이전트 매니저"),
            ("GS",  "cmd/modules/gs/",   "66",   "85",  "1,277",  "정보 수집 서비스"),
            ("AS",  "cmd/modules/as/",   "13",  "109",  "3,216",  "보안 스캔"),
            ("AH",  "cmd/modules/ah/",   "17",   "53",  "1,010",  "호스트 분석"),
            ("CC",  "cmd/cc/",           "17",   "55",  "1,287",  "컴플라이언스 체크"),
            ("CH",  "cmd/modules/ch/",   "78",   "57",  "1,070",  "보안 히스토리"),
        ],
        header_hex="1A237E",
    )
    doc.add_paragraph()

    # 4. SA 아키텍처
    h1("4. SA 모듈 아키텍처")
    body(
        "SA는 BoltDB 기반 Job 큐, gRPC 서버, HTTP 메시지 발신, "
        "로그 수집, 리소스 모니터링 등 10개 이상의 고루틴을 기동하여 "
        "전체 에이전트 시스템을 관리합니다."
    )
    doc.add_paragraph()

    h2("4.1 패키지 구조")
    styled_table(doc,
        ["패키지", "역할"],
        [
            ("moduleManager",        "모듈 생명주기 관리 (기동/종료/업데이트/상태)"),
            ("saJobManager",         "SA Job 큐 처리 (init/stop/configUpdate 등)"),
            ("messageSender",        "AM 서버 HTTP 통신 (Job 폴링, 결과 전송)"),
            ("grpcServer",           "모듈 → SA gRPC 수신 (HeartBeat, 완료 보고)"),
            ("dbManager",            "BoltDB 기반 로컬 Job 영속화"),
            ("logCollectionManager", "각 모듈 로그 파일 수집 및 관리"),
            ("resourceManager",      "CPU/MEM/DISK/Agent 리소스 모니터링"),
        ],
        header_hex="1B5E20",
    )
    doc.add_paragraph()

    h2("4.2 기동 흐름")
    steps = [
        ("1", "config.InitProcess",                         "설정 파일 로드, 로그 초기화"),
        ("2", "dbManager.InitDbManager",                    "BoltDB 오픈 + RoutineDbManager 기동"),
        ("3", "messageSender.InitMessageSender",            "HTTP 발신 루틴 기동"),
        ("4", "grpcServer.StartGrpcServer",                 "gRPC 서버 바인딩"),
        ("5", "moduleManager.InitModuleManager",            "모듈 관리 루틴 4개 기동"),
        ("6", "resourceManager.RoutineResourceManager",     "CPU/MEM/DISK 모니터 루틴 기동"),
        ("7", "logCollectionManager.RoutineLogCollectionManager", "로그 수집 루틴 기동"),
        ("8", "saJobManager.RoutineProcSaJobManager",       "SA Job 처리 루프 기동"),
        ("9", "wgSa.Wait",                                  "전체 고루틴 종료 대기"),
    ]
    styled_table(doc, ["#", "함수", "설명"], steps, header_hex="1B5E20")
    doc.add_paragraph()

    # 5. 기술 스택
    h1("5. 기술 스택")
    styled_table(doc,
        ["분류", "기술"],
        [
            ("언어",     "Go 1.21+"),
            ("통신",     "gRPC (protobuf), HTTP REST"),
            ("DB",       "BoltDB (로컬 파일 DB)"),
            ("빌드",     "Makefile / moduleBuild.sh"),
            ("서명",     "osslsigncode (Windows), smctl (SM인증서)"),
            ("CI/CD",    "GitLab CI"),
            ("문서/도구","Claude Code + indexer"),
        ],
        header_hex="4A148C",
    )
    # ── 문서 내용 끝 ──────────────────────────────────────────

    doc.save(output_path)
    return output_path


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("output", nargs="?", default=None)
    parser.add_argument("--data", default=None)
    args = parser.parse_args()

    data = {}
    if args.data and os.path.exists(args.data):
        with open(args.data, encoding="utf-8") as f:
            data = json.load(f)

    output_path = args.output or os.path.join(tempfile.gettempdir(), "output.docx")
    result = build(output_path, data)
    print(result)


if __name__ == "__main__":
    main()
