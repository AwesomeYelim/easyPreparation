// 시안 1 — 클래식 정통
// 와인/버건디 + 아이보리, 명조 강조, 격조있는 정통 스타일
// 장식: 얇은 더블 라인, 고전적 디바이더, 대문자 라틴 부제

const v1Styles = {
  paper: {
    width: 1200,
    height: 848, // ≈ 420×297mm 비율
    background: "#F5EFE4",
    color: "#2A1A1A",
    fontFamily: '"Noto Serif KR", "Nanum Myeongjo", serif',
    position: "relative",
    overflow: "hidden",
    display: "flex",
  },
  half: {
    flex: 1,
    padding: "56px 48px",
    position: "relative",
    display: "flex",
    flexDirection: "column",
  },
  divider: { borderLeft: "1px solid #C9B7A2" },
  wine: (typeof window !== 'undefined' && window.__BULLETIN_ACCENT__) || "#6B1F2A",
  ink: "#2A1A1A",
  cream: "#F5EFE4",
  rule: "#C9B7A2",
};

function V1DoubleRule({ thin = false }) {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 3, margin: "10px 0" }}>
      <div style={{ height: thin ? 0.5 : 1, background: v1Styles.wine }} />
      <div style={{ height: thin ? 0.5 : 1, background: v1Styles.wine }} />
    </div>
  );
}

function V1Cover({ data }) {
  return (
    <div style={{ ...v1Styles.half, justifyContent: "space-between", alignItems: "center", textAlign: "center" }}>
      <div style={{ width: "100%" }}>
        <div style={{ fontFamily: '"Cormorant Garamond", serif', fontSize: 11, letterSpacing: "0.4em", color: v1Styles.wine, textTransform: "uppercase", marginTop: 18 }}>
          ORDER OF SUNDAY WORSHIP
        </div>
        <V1DoubleRule />
        <div style={{ fontSize: 10, letterSpacing: "0.3em", color: v1Styles.ink, opacity: 0.7 }}>
          {data.churchNameEn}
        </div>
      </div>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", alignItems: "center", gap: 28 }}>
        {/* 표지 사진 또는 모노그램 */}
        {data.coverImage ? (
          <div style={{
            width: 220, height: 220,
            backgroundImage: `url(${data.coverImage})`,
            backgroundSize: "cover", backgroundPosition: "center",
            border: `1px solid ${v1Styles.wine}`,
            boxShadow: `0 0 0 6px ${v1Styles.cream}, 0 0 0 7px ${v1Styles.wine}`,
          }} />
        ) : (
          <div style={{
            width: 140, height: 140, borderRadius: "50%",
            border: `1px solid ${v1Styles.wine}`,
            display: "flex", alignItems: "center", justifyContent: "center",
            position: "relative",
          }}>
            <div style={{
              position: "absolute", inset: 8, borderRadius: "50%",
              border: `0.5px solid ${v1Styles.wine}`,
            }} />
            <div style={{ fontFamily: '"Cormorant Garamond", serif', fontSize: 56, color: v1Styles.wine, fontStyle: "italic", lineHeight: 1 }}>
              ☩
            </div>
          </div>
        )}

        <div>
          <div style={{ fontSize: 44, lineHeight: 1.25, color: v1Styles.ink, letterSpacing: "0.05em" }}>
            {data.churchName}
          </div>
          <div style={{ fontSize: 13, marginTop: 14, color: v1Styles.wine, letterSpacing: "0.5em" }}>
            주 일  예 배
          </div>
        </div>

        <div style={{ fontFamily: '"Cormorant Garamond", serif', fontSize: 18, color: v1Styles.ink, fontStyle: "italic", maxWidth: 380, lineHeight: 1.5 }}>
          “{data.verseQuote.text}”
          <div style={{ fontSize: 12, marginTop: 8, opacity: 0.6, fontStyle: "normal", letterSpacing: "0.1em" }}>
            — {data.verseQuote.ref}
          </div>
        </div>
      </div>

      <div style={{ width: "100%" }}>
        <V1DoubleRule />
        <div style={{ display: "flex", justifyContent: "space-between", fontSize: 11, letterSpacing: "0.2em", color: v1Styles.ink }}>
          <span>{data.date}</span>
          <span>{data.weekNumber}</span>
          <span>{data.pastor}</span>
        </div>
      </div>
    </div>
  );
}

function V1Back({ data }) {
  return (
    <div style={{ ...v1Styles.half, ...v1Styles.divider }}>
      <div style={{ fontFamily: '"Cormorant Garamond", serif', fontSize: 10, letterSpacing: "0.4em", color: v1Styles.wine, textTransform: "uppercase" }}>
        ANNOUNCEMENTS · 교회 소식
      </div>
      <V1DoubleRule />

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "18px 28px", marginTop: 12, fontSize: 11, lineHeight: 1.65 }}>
        {data.announcements.map((a, i) => (
          <div key={i}>
            <div style={{ fontSize: 13, color: v1Styles.wine, marginBottom: 4, letterSpacing: "0.05em" }}>
              {String(i + 1).padStart(2, "0")} · {a.title}
            </div>
            <div style={{ color: v1Styles.ink, opacity: 0.85 }}>{a.body}</div>
          </div>
        ))}
      </div>

      <div style={{ marginTop: "auto" }}>
        <V1DoubleRule thin />
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", fontSize: 10, color: v1Styles.ink, opacity: 0.7, letterSpacing: "0.1em" }}>
          <div>
            <div style={{ fontFamily: '"Cormorant Garamond", serif', fontSize: 16, fontStyle: "italic", color: v1Styles.wine, opacity: 1 }}>
              Soli Deo Gloria
            </div>
            <div style={{ marginTop: 4 }}>{data.tagline || "오직 하나님께만 영광을"}</div>
          </div>
          <div style={{ textAlign: "right" }}>
            <div>{data.website || ""}</div>
            <div style={{ marginTop: 2 }}>{data.blogInfo || ""}</div>
          </div>
        </div>
      </div>
    </div>
  );
}

function V1Inside({ data }) {
  return (
    <div style={v1Styles.paper}>
      {/* 좌면: 예배 순서 */}
      <div style={v1Styles.half}>
        <div style={{ fontFamily: '"Cormorant Garamond", serif', fontSize: 10, letterSpacing: "0.4em", color: v1Styles.wine, textTransform: "uppercase" }}>
          ORDER OF WORSHIP · 예 배 순 서
        </div>
        <V1DoubleRule />
        <div style={{ fontSize: 22, color: v1Styles.ink, marginTop: 4, letterSpacing: "0.1em" }}>
          오 전 예 배
        </div>
        <div style={{ fontSize: 11, color: v1Styles.ink, opacity: 0.6, marginTop: 4, letterSpacing: "0.1em" }}>
          오늘의 말씀 · {data.scripture}
        </div>

        <div style={{ marginTop: 18, fontSize: 12.5, lineHeight: 1.95 }}>
          {data.order.filter(o => !o.kind).map((o, i) => (
            <div key={i} style={{
              display: "grid",
              gridTemplateColumns: "84px 1fr 110px",
              alignItems: "baseline",
              borderBottom: `0.5px dotted ${v1Styles.rule}`,
              padding: "3px 0",
            }}>
              <div style={{ color: v1Styles.wine, letterSpacing: "0.15em", fontSize: 11.5 }}>
                {o.item}
              </div>
              <div style={{
                color: v1Styles.ink,
                fontStyle: o.emphasize ? "italic" : "normal",
                fontSize: o.emphasize ? 13 : 12,
                letterSpacing: "0.02em",
              }}>
                {o.content || "—"}
              </div>
              <div style={{ color: v1Styles.ink, opacity: 0.7, fontSize: 10.5, textAlign: "right", letterSpacing: "0.1em" }}>
                {o.who}
              </div>
            </div>
          ))}
        </div>

        <div style={{ marginTop: 16, fontSize: 10.5, color: v1Styles.ink, opacity: 0.75, letterSpacing: "0.1em" }}>
          {data.helpers.map(h => `${h.role}  ${h.name}`).join("    ·    ")}
        </div>
      </div>

      {/* 우면: 오후/수요 + 말씀 박스 */}
      <div style={{ ...v1Styles.half, ...v1Styles.divider }}>
        <div style={{ fontFamily: '"Cormorant Garamond", serif', fontSize: 10, letterSpacing: "0.4em", color: v1Styles.wine, textTransform: "uppercase" }}>
          AFTERNOON · MIDWEEK
        </div>
        <V1DoubleRule />

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 28, marginTop: 8 }}>
          <div>
            <div style={{ fontSize: 18, color: v1Styles.ink, letterSpacing: "0.1em" }}>오 후 예 배</div>
            <div style={{ marginTop: 14, fontSize: 12, lineHeight: 2.1 }}>
              {data.afternoon.map((o, i) => (
                <div key={i} style={{ display: "flex", justifyContent: "space-between", borderBottom: `0.5px dotted ${v1Styles.rule}` }}>
                  <span style={{ color: v1Styles.wine, letterSpacing: "0.15em" }}>{o.item}</span>
                  <span style={{ color: v1Styles.ink, opacity: 0.85 }}>{o.content || "—"}</span>
                </div>
              ))}
            </div>
          </div>
          <div>
            <div style={{ fontSize: 18, color: v1Styles.ink, letterSpacing: "0.1em" }}>수 요 예 배</div>
            <div style={{ marginTop: 14, fontSize: 12, lineHeight: 2.1 }}>
              {data.wednesday.map((o, i) => (
                <div key={i} style={{ display: "flex", justifyContent: "space-between", borderBottom: `0.5px dotted ${v1Styles.rule}` }}>
                  <span style={{ color: v1Styles.wine, letterSpacing: "0.15em" }}>{o.item}</span>
                  <span style={{ color: v1Styles.ink, opacity: 0.85 }}>{o.content || "—"}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* 말씀 강조 박스 — 남은 공간 전체를 채움 */}
        <div style={{
          marginTop: 24,
          flex: 1,
          padding: "28px 28px",
          border: `1px solid ${v1Styles.wine}`,
          background: "rgba(107,31,42,0.04)",
          position: "relative",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
        }}>
          <div style={{
            position: "absolute", top: -10, left: 24,
            background: v1Styles.cream, padding: "0 10px",
            fontFamily: '"Cormorant Garamond", serif', fontStyle: "italic",
            color: v1Styles.wine, fontSize: 14,
          }}>
            Verbum Dei
          </div>
          <div style={{ fontSize: 21, lineHeight: 1.55, color: v1Styles.ink, letterSpacing: "0.03em" }}>
            『{data.verseQuote.text}』
          </div>
          <div style={{ fontSize: 11, marginTop: 12, color: v1Styles.wine, letterSpacing: "0.2em", textAlign: "right" }}>
            — {data.verseQuote.ref}
          </div>
        </div>

        <div style={{ marginTop: 18 }}>
          <V1DoubleRule thin />
          <div style={{ fontSize: 10, color: v1Styles.ink, opacity: 0.6, textAlign: "center", letterSpacing: "0.3em" }}>
            ANNO · DOMINI · MMXXVI
          </div>
        </div>
      </div>
    </div>
  );
}

function V1Outside({ data }) {
  return (
    <div style={v1Styles.paper}>
      <V1Back data={data} />
      <V1Cover data={data} />
    </div>
  );
}

window.V1Outside = V1Outside;
window.V1Inside = V1Inside;
