// 시안 9 — 영수증/노트 컨셉
// 좁은 컬럼 + 모노스페이스 + 점선 perforation + 도트매트릭스 헤더
// 폴라로이드처럼 종이에 비스듬히 붙은 사진

const v9Styles = {
  paper: {
    width: 1200, height: 848,
    background: "#f5f1e8",
    color: "#1a1a17",
    display: "flex",
    fontFamily: "'Courier New', 'Roboto Mono', monospace",
    position: "relative",
    overflow: "hidden",
  },
  ink: "#1a1a17",
  inkLt: "#5a5650",
  paperBg: "#f5f1e8",
  rule: "#1a1a17",
  stamp: (typeof window !== 'undefined' && window.__BULLETIN_ACCENT__) || "#a83232",
  // 영수증 컬럼
  receipt: {
    width: 480,
    background: "#ffffff",
    marginTop: 30,
    padding: "32px 40px 40px",
    boxShadow: "0 8px 32px rgba(0,0,0,0.08), 0 0 0 1px rgba(0,0,0,0.04)",
    position: "relative",
    fontSize: 12,
    lineHeight: 1.7,
    letterSpacing: "0.02em",
  },
  noteSheet: {
    flex: 1,
    background: "#fbf6e8",
    backgroundImage: "repeating-linear-gradient(transparent 0, transparent 28px, rgba(26,26,23,0.12) 28px, rgba(26,26,23,0.12) 28.5px)",
    padding: "44px 48px 30px",
    position: "relative",
    fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
    overflow: "hidden",
  },
};

function parseDate(dateStr) {
  const parts = (dateStr || "").replace(/\s/g, "").split(".").filter(Boolean);
  const [y, m, d] = parts.map(Number);
  const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
  const monthsKo = ["1월","2월","3월","4월","5월","6월","7월","8월","9월","10월","11월","12월"];
  return { year: y, month: m, day: d, monthEn: months[(m||1)-1], monthKo: monthsKo[(m||1)-1] };
}
function weekNum(wn) { return (wn || "").replace(/[^0-9]/g, ""); }

// 점선 구분자 (영수증 가로 점선)
function V9Dashed({ char = "-", color }) {
  return (
    <div style={{
      letterSpacing: "0.1em", fontSize: 11,
      color: color || v9Styles.inkLt,
      whiteSpace: "nowrap", overflow: "hidden",
      userSelect: "none",
    }}>
      {char.repeat(80)}
    </div>
  );
}

// 영수증 위/아래 톱니 가장자리
function V9Zigzag({ flip = false }) {
  return (
    <div style={{
      height: 12,
      background: `linear-gradient(135deg, ${v9Styles.paperBg} 33%, #ffffff 33%) 0 0/12px 24px,
                   linear-gradient(225deg, ${v9Styles.paperBg} 33%, #ffffff 33%) 12px 0/12px 24px`,
      transform: flip ? "scaleY(-1)" : "none",
    }} />
  );
}

// 줄 (라벨 ............ 값) — 영수증 dot leader
function V9Line({ label, value, bold, big }) {
  const dotsCount = Math.max(2, 38 - label.length * 2 - String(value || "").length);
  return (
    <div style={{
      display: "flex", justifyContent: "space-between",
      fontSize: big ? 13 : 12, fontWeight: bold ? 700 : 400,
      color: v9Styles.ink,
      padding: "1px 0",
    }}>
      <span style={{ textTransform: bold ? "uppercase" : "none", letterSpacing: bold ? "0.1em" : "0.02em" }}>
        {label}
      </span>
      <span style={{ flex: 1, color: v9Styles.inkLt, padding: "0 4px", overflow: "hidden", letterSpacing: 1 }}>
        {".".repeat(dotsCount)}
      </span>
      <span style={{ fontVariantNumeric: "tabular-nums" }}>{value}</span>
    </div>
  );
}

function V9Cover({ data }) {
  const orderItems = data.order.filter(o => !o.kind);
  return (
    <div style={{
      flex: 1,
      background: "#ffffff",
      display: "flex",
      flexDirection: "column",
      overflow: "hidden",
      position: "relative",
      fontFamily: "'Courier New', 'Roboto Mono', monospace",
      fontSize: 12,
      lineHeight: 1.7,
      letterSpacing: "0.02em",
      color: v9Styles.ink,
    }}>
      {/* 워터마크 도장 */}
      <div style={{
        position: "absolute", right: 40, top: 60,
        width: 130, height: 130,
        border: `3px solid ${v9Styles.stamp}`,
        borderRadius: "50%",
        display: "flex", flexDirection: "column",
        alignItems: "center", justifyContent: "center",
        transform: "rotate(-12deg)",
        opacity: 0.85,
        color: v9Styles.stamp,
        fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
        zIndex: 2,
      }}>
        <div style={{ fontSize: 11, letterSpacing: "0.4em", borderTop: `1px solid ${v9Styles.stamp}`, borderBottom: `1px solid ${v9Styles.stamp}`, padding: "2px 0", width: "70%", textAlign: "center" }}>
          접 수
        </div>
        <div style={{ fontSize: 18, fontWeight: 700, marginTop: 6, letterSpacing: "0.05em" }}>
          {`제${weekNum(data.weekNumber)}주`}
        </div>
        <div style={{ fontSize: 9, letterSpacing: "0.2em", marginTop: 4 }}>
          {`${String(parseDate(data.date).day).padStart(2,'0')} ${parseDate(data.date).monthEn.toUpperCase()} ${parseDate(data.date).year}`}
        </div>
      </div>

      <V9Zigzag />
      <div style={{ flex: 1, minHeight: 0, padding: "18px 44px 8px", overflow: "hidden", display: "flex", flexDirection: "column" }}>
        {/* 도트매트릭스 헤더 */}
        <div style={{ textAlign: "center", letterSpacing: "0.5em", fontSize: 10, paddingBottom: 4 }}>
          ▓ ▓ ▓  주  보  ▓ ▓ ▓
        </div>
        <div style={{
          textAlign: "center",
          fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
          fontSize: 32, fontWeight: 700, letterSpacing: "0.05em",
          lineHeight: 1.2, padding: "8px 0",
        }}>
          {data.churchName}
        </div>
        <div style={{ textAlign: "center", fontSize: 10, letterSpacing: "0.25em", color: v9Styles.inkLt }}>
          {data.churchNameEn}
        </div>

        <div style={{ marginTop: 14 }}><V9Dashed char="=" /></div>

        {/* 메타 */}
        <div style={{ padding: "10px 0", fontSize: 11.5 }}>
          <V9Line label="발 행 일" value={data.date} />
          <V9Line label="주    차" value={data.weekNumber} />
          <V9Line label="담 임" value={data.pastor} />
        </div>

        <V9Dashed />

        {/* 사진 — 폴라로이드 */}
        <div style={{
          margin: "8px auto",
          padding: 8, paddingBottom: 26,
          background: "#fefefe",
          border: "1px solid rgba(0,0,0,0.08)",
          boxShadow: "0 4px 14px rgba(0,0,0,0.18), 0 1px 3px rgba(0,0,0,0.1)",
          transform: "rotate(-1.5deg)",
          width: 260,
          position: "relative",
        }}>
          {data.coverImage ? (
            <div style={{
              width: "100%", height: 150,
              backgroundImage: `url(${data.coverImage})`,
              backgroundSize: "cover", backgroundPosition: "center",
            }} />
          ) : (
            <div style={{
              width: "100%", height: 150,
              background: "repeating-linear-gradient(45deg, #ebe6d8 0 8px, #e2dccb 8px 16px)",
              display: "flex", alignItems: "center", justifyContent: "center",
              color: v9Styles.inkLt, fontSize: 10, letterSpacing: "0.3em",
            }}>
              사진 자리 · PHOTO
            </div>
          )}
          {/* 마스킹 테이프 */}
          <div style={{
            position: "absolute", top: -10, left: 24,
            width: 60, height: 18,
            background: "rgba(220, 200, 100, 0.55)",
            transform: "rotate(-4deg)",
            boxShadow: "0 1px 3px rgba(0,0,0,0.06)",
          }} />
          <div style={{
            position: "absolute", bottom: 6, left: 0, right: 0,
            fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
            fontSize: 12, textAlign: "center",
            fontStyle: "italic", color: v9Styles.inkLt,
          }}>
            {data.date}
          </div>
        </div>

        <V9Dashed />

        {/* 오늘의 설교 — 영수증 항목 */}
        <div style={{ padding: "8px 0", fontSize: 11.5 }}>
          <div style={{ fontWeight: 700, letterSpacing: "0.15em", textAlign: "center", padding: "4px 0" }}>
            ▌ 금주 설교 ▌
          </div>
          <V9Line label="제 목" value={data.sermonTitle} bold />
          <V9Line label="본 문" value={data.scripture} />
        </div>

        <V9Dashed char="=" />

        {/* 인용문 — flex:1로 남은 공간 채움 */}
        <div style={{
          flex: 1,
          display: "flex", flexDirection: "column", justifyContent: "center",
          padding: "8px 4px",
          fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
          fontSize: 14.5, lineHeight: 1.7,
          textAlign: "center",
          letterSpacing: "0.02em", wordBreak: "keep-all",
        }}>
          「{data.verseQuote.text}」
          <div style={{ fontSize: 10.5, color: v9Styles.inkLt, letterSpacing: "0.15em", marginTop: 6 }}>
            ─ {data.verseQuote.ref} ─
          </div>
        </div>
      </div>

      {/* 고정 하단 푸터 — flexShrink:0으로 항상 하단에 붙음 */}
      <div style={{ flexShrink: 0, padding: "0 44px 6px" }}>
        <V9Dashed char="=" />
        <div style={{ textAlign: "center", padding: "6px 0", fontSize: 9.5, color: v9Styles.inkLt, letterSpacing: "0.2em" }}>
          ※ 오전 {orderItems.length}개 순서 진행 ※
          <br />
          감사합니다 · THANK YOU
        </div>
        <div style={{ margin: "6px 24px 0" }}>
          <div style={{
            height: 32,
            background: `repeating-linear-gradient(90deg, ${v9Styles.ink} 0 2px, transparent 2px 4px, ${v9Styles.ink} 4px 5px, transparent 5px 9px, ${v9Styles.ink} 9px 12px, transparent 12px 14px)`,
          }} />
          <div style={{ fontSize: 9, letterSpacing: "0.25em", textAlign: "center", marginTop: 2 }}>
            {`0190 ${weekNum(data.weekNumber)} ${String(parseDate(data.date).day).padStart(2,'0')} ${String(parseDate(data.date).month).padStart(2,'0')} ${parseDate(data.date).year} LLC`}
          </div>
        </div>
      </div>
      <V9Zigzag flip />
    </div>
  );
}

function V9Back({ data }) {
  return (
    <div style={v9Styles.noteSheet}>
      {/* 노트 상단 헤더 */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", marginBottom: 6 }}>
        <div style={{
          fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
          fontSize: 30, fontWeight: 700, letterSpacing: "0.02em",
        }}>
          교회 소식 메모
        </div>
        <div style={{
          fontFamily: "'Courier New', monospace",
          fontSize: 11, color: v9Styles.inkLt, letterSpacing: "0.2em",
        }}>
          NO. {data.weekNumber.replace(/[^0-9]/g, "")}
        </div>
      </div>
      <div style={{ height: 2, background: v9Styles.ink, marginBottom: 14 }} />

      {/* 손글씨 풍 본문 */}
      <div style={{ fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif", fontSize: data.announcements.length > 10 ? "82%" : data.announcements.length > 7 ? "90%" : "100%" }}>
        {data.announcements.map((a, i) => (
          <div key={i} style={{
            marginBottom: data.announcements.length > 10 ? 5 : data.announcements.length > 7 ? 8 : 14,
            paddingBottom: data.announcements.length > 10 ? 4 : data.announcements.length > 7 ? 6 : 10,
            borderBottom: `1px dashed rgba(26,26,23,0.3)`,
          }}>
            <div style={{ display: "flex", alignItems: "baseline", gap: 10 }}>
              <span style={{
                fontFamily: "'Courier New', monospace",
                fontSize: 11, fontWeight: 700,
                background: v9Styles.stamp, color: "#fff",
                padding: "2px 7px",
                letterSpacing: "0.1em",
              }}>
                {String(i + 1).padStart(2, "0")}
              </span>
              <span style={{ fontSize: 16, fontWeight: 700, wordBreak: "keep-all" }}>{a.title}</span>
            </div>
            <div style={{
              fontSize: 13, lineHeight: 1.65, marginTop: 6,
              color: v9Styles.ink, paddingLeft: 4, wordBreak: "keep-all",
            }}>
              {a.body}
            </div>
          </div>
        ))}

        {/* 하단 손글씨 메모 */}
        <div style={{
          marginTop: 18,
          fontFamily: "'Courier New', monospace",
          fontSize: 11, letterSpacing: "0.1em",
          color: v9Styles.inkLt,
        }}>
          ▶ {data.tagline || "오직 하나님께만 영광을"} · {data.pastor}
        </div>
      </div>

      {/* 우상단 종이 클립(미니멀 SVG 대신 모양만) */}
      <div style={{
        position: "absolute", top: 24, right: 28,
        width: 28, height: 60,
        border: `2.5px solid ${v9Styles.ink}`,
        borderRadius: "12px 12px 12px 12px",
        borderBottom: "none",
        opacity: 0.4,
      }} />
    </div>
  );
}

function V9Inside({ data }) {
  const orderItems = data.order.filter(o => !o.kind);
  return (
    <div style={v9Styles.paper}>
      {/* 좌측 — 예배 순서 (열 전체를 흰 영수증 배경으로 채움) */}
      <div style={{
        flex: 1,
        background: "#ffffff",
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
        position: "relative",
        boxShadow: "2px 0 12px rgba(0,0,0,0.06)",
        fontFamily: "'Courier New', 'Roboto Mono', monospace",
        fontSize: 12,
        lineHeight: 1.7,
        letterSpacing: "0.02em",
        color: v9Styles.ink,
      }}>
        {/* 상단 zigzag */}
        <V9Zigzag />
        <div style={{ flex: 1, minHeight: 0, padding: "18px 44px 8px", overflow: "hidden", display: "flex", flexDirection: "column" }}>
          <div style={{ textAlign: "center", letterSpacing: "0.5em", fontSize: 10 }}>
            ▓ ▓  영  수  증  ▓ ▓
          </div>
          <div style={{
            textAlign: "center",
            fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
            fontSize: 26, fontWeight: 700, padding: "8px 0",
          }}>
            오전 예배 순서
          </div>
          <div style={{ textAlign: "center", fontSize: 10, letterSpacing: "0.25em", color: v9Styles.inkLt }}>
            {data.date} · {data.scripture}
          </div>

          <div style={{ marginTop: 14 }}><V9Dashed char="=" /></div>

          {/* 항목 */}
          <div style={{ padding: "10px 0", fontSize: 12 }}>
            {orderItems.map((o, i) => (
              <div key={i} style={{ padding: "1px 0" }}>
                <div style={{
                  display: "grid",
                  gridTemplateColumns: "28px 1fr auto",
                  gap: 6,
                  fontWeight: o.emphasize ? 700 : 400,
                  background: o.emphasize ? "rgba(168,50,50,0.07)" : "transparent",
                }}>
                  <span style={{ fontVariantNumeric: "tabular-nums" }}>
                    {String(i + 1).padStart(2, "0")}.
                  </span>
                  <span style={{ overflow: "hidden", whiteSpace: "nowrap", textOverflow: "ellipsis" }}>
                    {o.item}
                  </span>
                  <span style={{ color: v9Styles.inkLt, fontSize: 11, letterSpacing: "0.05em" }}>
                    {o.who}
                  </span>
                </div>
                {o.content && (
                  <div style={{
                    paddingLeft: 34,
                    fontSize: 11, color: v9Styles.inkLt,
                    fontFamily: o.emphasize ? "'Nanum Myeongjo', serif" : "inherit",
                    fontStyle: o.emphasize ? "italic" : "normal",
                  }}>
                    └ {o.content}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* 고정 하단 푸터 */}
        <div style={{ flexShrink: 0, padding: "0 44px 6px" }}>
          <V9Dashed />
          <div style={{ padding: "8px 0", fontSize: 11 }}>
            <div style={{ fontWeight: 700, letterSpacing: "0.15em", padding: "4px 0" }}>
              ▌ 봉 사 자
            </div>
            {data.helpers.map((h, i) => (
              <V9Line key={i} label={h.role} value={h.name} />
            ))}
          </div>
          <V9Dashed char="=" />
          <div style={{ padding: "6px 0", textAlign: "center", fontSize: 11 }}>
            <div style={{ fontWeight: 700, letterSpacing: "0.15em" }}>
              * * 합  계 * *
            </div>
            <div style={{ fontSize: 14, fontWeight: 700, marginTop: 4 }}>
              예배 순서 {orderItems.length} 가지
            </div>
          </div>
          <V9Dashed char="=" />
          <div style={{ textAlign: "center", padding: "6px 0", fontSize: 9.5, color: v9Styles.inkLt, letterSpacing: "0.2em" }}>
            {`은 혜 안 에 서 . ${(parseDate(data.date).year || new Date().getFullYear()).toString().split('').join(' ')}`}
          </div>
        </div>
        <V9Zigzag flip />
      </div>

      {/* 우측 노트 — 묵상 + 그 외 예배 */}
      <div style={{ ...v9Styles.noteSheet, flex: 1, borderLeft: `1.5px dashed ${v9Styles.ink}` }}>
        {/* 손으로 쓴 듯한 제목 */}
        <div style={{
          fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
          fontSize: 38, fontWeight: 700,
          letterSpacing: "0.02em",
          marginBottom: 4,
        }}>
          오늘의 묵상
        </div>
        <div style={{
          fontFamily: "'Courier New', monospace",
          fontSize: 11, color: v9Styles.inkLt, letterSpacing: "0.2em",
          marginBottom: 14,
        }}>
          {data.scripture}
        </div>

        {/* 큰 인용문 */}
        <div style={{
          fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
          fontSize: 22, lineHeight: 1.75, fontWeight: 500,
          padding: "12px 0", marginBottom: 8, wordBreak: "keep-all",
        }}>
          「{data.verseQuote.text}」
        </div>
        <div style={{
          fontFamily: "'Courier New', monospace",
          fontSize: 12, letterSpacing: "0.15em",
          color: v9Styles.stamp, marginBottom: 18,
        }}>
          → {data.verseQuote.ref}
        </div>

        <div style={{ height: 2, background: v9Styles.ink, marginBottom: 14, marginTop: 6 }} />

        {/* 오후/수요 — 두 개 메모지 */}
        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          {[
            { title: "오후 예배", items: data.afternoon },
            { title: "수요 예배", items: data.wednesday },
          ].map((sec, idx) => (
            <div key={idx} style={{
              fontFamily: "'Nanum Myeongjo', 'Noto Serif KR', serif",
            }}>
              <div style={{ display: "flex", alignItems: "baseline", gap: 10 }}>
                <span style={{
                  fontFamily: "'Courier New', monospace",
                  fontSize: 11, fontWeight: 700,
                  background: v9Styles.ink, color: v9Styles.paperBg,
                  padding: "2px 8px", letterSpacing: "0.15em",
                }}>
                  {String.fromCharCode(65 + idx)}
                </span>
                <span style={{ fontSize: 20, fontWeight: 700 }}>{sec.title}</span>
              </div>
              <div style={{ marginTop: 6, fontSize: 12.5, lineHeight: 1.8 }}>
                {sec.items.map((o, j) => (
                  <div key={j} style={{ display: "flex", justifyContent: "space-between", borderBottom: `1px dotted rgba(26,26,23,0.3)`, padding: "1px 0" }}>
                    <span>{o.item}</span>
                    <span style={{ color: v9Styles.inkLt }}>{o.content || "—"}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* 우하단 손글씨 메모 */}
        <div style={{
          marginTop: 18, paddingTop: 8,
          borderTop: `1px solid rgba(26,26,23,0.4)`,
          fontFamily: "'Courier New', monospace",
          fontSize: 10.5, letterSpacing: "0.15em",
          color: v9Styles.inkLt,
          display: "flex", justifyContent: "space-between",
        }}>
          <span>p. {data.churchName}</span>
          <span>{data.date}</span>
        </div>
      </div>
    </div>
  );
}

function V9Outside({ data }) {
  return (
    <div style={v9Styles.paper}>
      <V9Back data={data} />
      <V9Cover data={data} />
    </div>
  );
}

window.V9Outside = V9Outside;
window.V9Inside = V9Inside;
