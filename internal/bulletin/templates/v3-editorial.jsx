// 시안 3 — 에디토리얼 (잡지형)
// 큰 숫자 타이포그래피 + 비대칭 레이아웃 + 풍부한 그리드
// 차분한 네이비 + 크림 + 와인 액센트

const v3Styles = {
  paper: {
    width: 1200,
    height: 848,
    background: "#EFE9DD",
    color: "#1A1815",
    fontFamily: '"Noto Sans KR", "Inter", system-ui, sans-serif',
    position: "relative",
    overflow: "hidden",
    display: "flex",
  },
  half: { flex: 1, padding: "48px 44px", position: "relative", display: "flex", flexDirection: "column" },
  ink: "#1A1815",
  cream: "#EFE9DD",
  accent: (typeof window !== 'undefined' && window.__BULLETIN_ACCENT__) || "#7A2C2C",
  navy: "#1F2E3D",
  rule: "#B6A992",
  serif: '"Noto Serif KR", "Nanum Myeongjo", serif',
};

function parseDate(dateStr) {
  const parts = (dateStr || "").replace(/\s/g, "").split(".").filter(Boolean);
  const [y, m, d] = parts.map(Number);
  const months = ["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"];
  const monthsKo = ["1월","2월","3월","4월","5월","6월","7월","8월","9월","10월","11월","12월"];
  return { year: y, month: m, day: d, monthEn: months[(m||1)-1], monthKo: monthsKo[(m||1)-1] };
}
function weekNum(wn) { return (wn || "").replace(/[^0-9]/g, ""); }

function V3Cover({ data }) {
  return (
    <div style={{ ...v3Styles.half, padding: 0, display: "block" }}>
      {/* 상단 매스헤드 */}
      <div style={{ padding: "20px 44px", borderBottom: `2px solid ${v3Styles.ink}`, display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
        <div style={{ fontFamily: v3Styles.serif, fontStyle: "italic", fontSize: 16, color: v3Styles.ink }}>
          주일자
        </div>
        <div style={{ fontSize: 9, letterSpacing: "0.3em", color: v3Styles.ink, opacity: 0.7 }}>
          {`${weekNum(data.weekNumber)}호 · 통권${weekNum(data.weekNumber)}호`}
        </div>
      </div>

      {/* 거대한 숫자 + 제목 — 사진 있으면 사진으로 교체 */}
      {data.coverImage ? (
        <div style={{
          margin: "0", height: 320,
          backgroundImage: `url(${data.coverImage})`,
          backgroundSize: "cover", backgroundPosition: "center",
          position: "relative",
          borderBottom: `1px solid ${v3Styles.rule}`,
        }}>
          {/* 좌하단 큰 17 오버레이 */}
          <div style={{
            position: "absolute", left: 32, bottom: 14,
            display: "flex", alignItems: "flex-end", gap: 12,
          }}>
            <div style={{
              fontFamily: v3Styles.serif, fontSize: 180, lineHeight: 0.78,
              color: v3Styles.cream, letterSpacing: "-0.06em", fontWeight: 400,
              textShadow: "0 2px 18px rgba(0,0,0,0.35)",
            }}>
              {weekNum(data.weekNumber)}
            </div>
            <div style={{
              fontFamily: v3Styles.serif, fontStyle: "italic",
              fontSize: 26, lineHeight: 1.1, color: v3Styles.cream,
              paddingBottom: 22, textShadow: "0 1px 8px rgba(0,0,0,0.45)",
            }}>
              주차
            </div>
          </div>
          {/* 우상단 캡션 */}
          <div style={{
            position: "absolute", top: 14, right: 18,
            background: v3Styles.cream, color: v3Styles.ink,
            fontFamily: v3Styles.serif, fontStyle: "italic", fontSize: 12,
            padding: "5px 10px", letterSpacing: "0.05em",
          }}>
            {`표지 · ${data.date.replace(/\s/g, '')}`}
          </div>
        </div>
      ) : (
        <>
          <div style={{ padding: "36px 44px 0", position: "relative", display: "flex", alignItems: "flex-end", gap: 18 }}>
            <div style={{ fontFamily: v3Styles.serif, fontSize: 280, lineHeight: 0.78, color: v3Styles.accent, letterSpacing: "-0.06em", fontWeight: 400 }}>
              {weekNum(data.weekNumber)}
            </div>
            <div style={{
              fontFamily: v3Styles.serif, fontStyle: "italic",
              fontSize: 38, lineHeight: 1.1, color: v3Styles.ink,
              paddingBottom: 30,
            }}>
              주차
            </div>
          </div>
        </>
      )}

      <div style={{ padding: "16px 44px 24px", borderBottom: `1px solid ${v3Styles.rule}` }}>
        <div style={{ fontSize: 11, letterSpacing: "0.35em", color: v3Styles.ink, opacity: 0.7 }}>
          {`${data.weekNumber} · ${parseDate(data.date).monthKo} ${parseDate(data.date).day}일 주일판`}
        </div>
      </div>

      {/* 제호 */}
      <div style={{ padding: "32px 44px 24px" }}>
        <div style={{ fontSize: 11, letterSpacing: "0.5em", color: v3Styles.accent }}>
          {`——  ${data.tagline || "은혜 가운데서"}  ——`}
        </div>
        <div style={{ fontFamily: v3Styles.serif, fontSize: 56, lineHeight: 1.05, color: v3Styles.ink, marginTop: 14, letterSpacing: "-0.01em" }}>
          {data.churchName}
        </div>
        <div style={{ marginTop: 18, fontFamily: v3Styles.serif, fontStyle: "italic", fontSize: 18, color: v3Styles.ink, maxWidth: 420, lineHeight: 1.55 }}>
          “{data.verseQuote.text}”
          <span style={{ fontStyle: "normal", fontSize: 11, marginLeft: 10, color: v3Styles.accent, letterSpacing: "0.15em" }}>
            {data.verseQuote.ref}
          </span>
        </div>
      </div>

      <div style={{ padding: "20px 44px", borderTop: `1px solid ${v3Styles.rule}`, display: "flex", justifyContent: "space-between", fontSize: 10, letterSpacing: "0.2em", color: v3Styles.ink, opacity: 0.7 }}>
        <span>{`${data.date.replace(/\s/g,'').replace(/\.$/,'')} 일`}</span>
        <span>설교 · {data.sermonTitle}</span>
        <span>{data.pastor}</span>
      </div>
    </div>
  );
}

function V3Back({ data }) {
  return (
    <div style={{ ...v3Styles.half, borderLeft: `1px solid ${v3Styles.rule}`, padding: 0 }}>
      <div style={{ padding: "20px 44px", borderBottom: `2px solid ${v3Styles.ink}`, display: "flex", justifyContent: "space-between" }}>
        <div style={{ fontFamily: v3Styles.serif, fontStyle: "italic", fontSize: 16 }}>교회 소식</div>
        <div style={{ fontSize: 9, letterSpacing: "0.3em", opacity: 0.7 }}>교회 소식</div>
      </div>

      <div style={{ padding: "24px 44px", flex: 1, display: "flex", flexDirection: "column", gap: 0 }}>
        {data.announcements.map((a, i) => (
          <div key={i} style={{
            display: "grid",
            gridTemplateColumns: "44px 1fr",
            gap: 14,
            padding: "10px 0",
            borderBottom: i === data.announcements.length - 1 ? "none" : `0.5px solid ${v3Styles.rule}`,
          }}>
            <div style={{
              fontFamily: v3Styles.serif, fontSize: 28,
              color: v3Styles.accent, lineHeight: 1, letterSpacing: "-0.02em",
              fontVariantNumeric: "tabular-nums",
            }}>
              {String(i + 1).padStart(2, "0")}
            </div>
            <div>
              <div style={{ fontSize: 13, color: v3Styles.ink, fontWeight: 600, letterSpacing: "0.02em" }}>
                {a.title}
              </div>
              <div style={{ fontSize: 11, color: v3Styles.ink, opacity: 0.78, lineHeight: 1.55, marginTop: 2 }}>
                {a.body}
              </div>
            </div>
          </div>
        ))}
      </div>

      <div style={{ padding: "16px 44px", borderTop: `2px solid ${v3Styles.ink}`, display: "flex", justifyContent: "space-between", fontSize: 10, letterSpacing: "0.2em", opacity: 0.7 }}>
        <span>{data.website || ""}</span>
        <span style={{ fontFamily: v3Styles.serif, fontStyle: "italic", opacity: 1, color: v3Styles.accent }}>
          Soli Deo Gloria
        </span>
        <span>{data.blogInfo || ""}</span>
      </div>
    </div>
  );
}

function V3Inside({ data }) {
  return (
    <div style={v3Styles.paper}>
      <div style={{ ...v3Styles.half, padding: 0 }}>
        <div style={{ padding: "20px 44px", borderBottom: `2px solid ${v3Styles.ink}`, display: "flex", justifyContent: "space-between" }}>
          <div style={{ fontFamily: v3Styles.serif, fontStyle: "italic", fontSize: 16 }}>예배 순서</div>
          <div style={{ fontSize: 9, letterSpacing: "0.3em", opacity: 0.7 }}>예 배 순 서</div>
        </div>

        <div style={{ padding: "24px 44px 0", display: "flex", alignItems: "flex-end", gap: 16 }}>
          <div style={{ fontFamily: v3Styles.serif, fontSize: 80, lineHeight: 0.9, color: v3Styles.ink, letterSpacing: "-0.03em" }}>
            오전
          </div>
          <div style={{ fontFamily: v3Styles.serif, fontStyle: "italic", fontSize: 30, color: v3Styles.accent, paddingBottom: 8 }}>
            예배
          </div>
        </div>
        <div style={{ padding: "6px 44px 14px", fontSize: 11, letterSpacing: "0.2em", color: v3Styles.ink, opacity: 0.7 }}>
          오늘의 본문 · {data.scripture}
        </div>

        <div style={{ padding: "0 44px", fontSize: 12, lineHeight: 1.95 }}>
          {data.order.filter(o => !o.kind).map((o, i) => (
            <div key={i} style={{
              display: "grid",
              gridTemplateColumns: "26px 92px 1fr 100px",
              gap: 10,
              alignItems: "baseline",
              padding: "3px 0",
              borderBottom: `0.5px dotted ${v3Styles.rule}`,
            }}>
              <span style={{ fontSize: 9, color: v3Styles.accent, fontVariantNumeric: "tabular-nums" }}>
                {String(i + 1).padStart(2, "0")}
              </span>
              <span style={{ fontFamily: v3Styles.serif, fontSize: 12.5, color: v3Styles.ink, letterSpacing: "0.05em" }}>
                {o.item}
              </span>
              <span style={{
                color: v3Styles.ink, opacity: 0.85,
                fontFamily: o.emphasize ? v3Styles.serif : "inherit",
                fontStyle: o.emphasize ? "italic" : "normal",
                fontSize: o.emphasize ? 13.5 : 11.5,
              }}>
                {o.content || "—"}
              </span>
              <span style={{ fontSize: 10, color: v3Styles.ink, opacity: 0.6, textAlign: "right", letterSpacing: "0.05em" }}>
                {o.who}
              </span>
            </div>
          ))}
        </div>

        <div style={{ padding: "12px 44px", fontSize: 10, color: v3Styles.ink, opacity: 0.65, letterSpacing: "0.1em" }}>
          {data.helpers.map(h => `${h.role} ${h.name}`).join("   ·   ")}
        </div>
      </div>

      <div style={{ ...v3Styles.half, padding: 0, borderLeft: `1px solid ${v3Styles.rule}` }}>
        {/* 풀블리드 말씀 박스 */}
        <div style={{ background: v3Styles.navy, color: v3Styles.cream, padding: "36px 44px", position: "relative" }}>
          <div style={{ fontSize: 9, letterSpacing: "0.4em", opacity: 0.7 }}>오늘의 말씀</div>
          <div style={{ fontFamily: v3Styles.serif, fontSize: 28, lineHeight: 1.5, marginTop: 14, letterSpacing: "0.01em" }}>
            “{data.verseQuote.text}”
          </div>
          <div style={{ fontSize: 11, letterSpacing: "0.3em", marginTop: 16, opacity: 0.85 }}>
            — {data.verseQuote.ref}
          </div>
        </div>

        {/* 오후 + 수요 */}
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 0, flex: 1 }}>
          {[
            { title: "오후 예배", en: "오후", items: data.afternoon, accent: false },
            { title: "수요 예배", en: "수요", items: data.wednesday, accent: true },
          ].map((sec, idx) => (
            <div key={idx} style={{
              padding: "26px 28px",
              borderRight: idx === 0 ? `1px solid ${v3Styles.rule}` : "none",
              background: sec.accent ? "rgba(122,44,44,0.06)" : "transparent",
            }}>
              <div style={{ fontFamily: v3Styles.serif, fontStyle: "italic", fontSize: 9, letterSpacing: "0.3em", color: v3Styles.accent }}>
                {sec.en}
              </div>
              <div style={{ fontFamily: v3Styles.serif, fontSize: 26, color: v3Styles.ink, marginTop: 4 }}>
                {sec.title}
              </div>
              <div style={{ marginTop: 14, fontSize: 11.5, lineHeight: 2 }}>
                {sec.items.map((o, i) => (
                  <div key={i} style={{ display: "flex", justifyContent: "space-between", borderBottom: `0.5px dotted ${v3Styles.rule}` }}>
                    <span style={{ fontFamily: v3Styles.serif, color: v3Styles.ink }}>{o.item}</span>
                    <span style={{ color: v3Styles.ink, opacity: 0.75 }}>{o.content || "—"}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div style={{ padding: "14px 44px", borderTop: `2px solid ${v3Styles.ink}`, display: "flex", justifyContent: "space-between", fontSize: 10, letterSpacing: "0.2em", opacity: 0.7 }}>
          <span>{data.churchName}</span>
          <span>P. {data.pastor}</span>
        </div>
      </div>
    </div>
  );
}

function V3Outside({ data }) {
  return (
    <div style={v3Styles.paper}>
      <V3Back data={data} />
      <V3Cover data={data} />
    </div>
  );
}

window.V3Outside = V3Outside;
window.V3Inside = V3Inside;
