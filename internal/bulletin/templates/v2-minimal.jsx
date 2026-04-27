// 시안 2 — 미니멀 모던
// 차분한 네이비/딥블루 + 오프화이트, 그리드 중심, 큰 여백
// 장식: 가는 헤어라인, 큰 숫자(주차), 정렬된 그리드

const v2Styles = {
  paper: {
    width: 1200,
    height: 848,
    background: "#FAFAF7",
    color: "#0F1B2D",
    fontFamily: '"Inter", "Noto Sans KR", system-ui, sans-serif',
    position: "relative",
    overflow: "hidden",
    display: "flex",
  },
  half: {
    flex: 1,
    padding: "64px 56px",
    position: "relative",
    display: "flex",
    flexDirection: "column",
  },
  divider: { borderLeft: "1px solid #DDDAD0" },
  navy: "#0F1B2D",
  ink: "#1A2230",
  bg: "#FAFAF7",
  rule: "#DDDAD0",
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

function V2Hairline({ vertical = false, length = "100%" }) {
  return <div style={{ background: v2Styles.rule, ...(vertical ? { width: 1, height: length } : { height: 1, width: length }) }} />;
}

function V2Cover({ data }) {
  return (
    <div style={{ ...v2Styles.half, justifyContent: "space-between" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div style={{ fontSize: 10, letterSpacing: "0.3em", color: v2Styles.navy, textTransform: "uppercase" }}>
          Sunday Worship
        </div>
        <div style={{ fontSize: 10, letterSpacing: "0.3em", color: v2Styles.ink, opacity: 0.5 }}>
          № {data.weekNumber.replace(/[^0-9]/g, "")} / 52
        </div>
      </div>

      {/* 거대한 날짜 타이포 */}
      <div style={{ marginTop: 40 }}>
        <div style={{ fontFamily: v2Styles.serif, fontSize: 220, lineHeight: 0.85, color: v2Styles.navy, letterSpacing: "-0.04em" }}>
          {parseDate(data.date).day}
        </div>
        <div style={{ display: "flex", alignItems: "baseline", gap: 16, marginTop: 12 }}>
          <span style={{ fontFamily: v2Styles.serif, fontSize: 32, color: v2Styles.navy }}>{parseDate(data.date).monthEn}</span>
          <span style={{ fontSize: 13, letterSpacing: "0.3em", color: v2Styles.ink, opacity: 0.6 }}>
            {parseDate(data.date).year} · SUNDAY
          </span>
        </div>
      </div>

      {/* 표지 이미지 or 여백 */}
      {data.coverImage ? (
        <div style={{
          flex: 1,
          marginTop: 32,
          backgroundImage: `url(${data.coverImage})`,
          backgroundSize: "cover",
          backgroundPosition: "center",
          maxHeight: 220,
        }} />
      ) : (
        <div style={{ flex: 1 }} />
      )}

      <div>
        <V2Hairline />
        <div style={{ marginTop: 20, fontSize: 11, letterSpacing: "0.4em", color: v2Styles.navy, textTransform: "uppercase", opacity: 0.7 }}>
          {data.churchNameEn}
        </div>
        <div style={{ fontFamily: v2Styles.serif, fontSize: 32, color: v2Styles.navy, marginTop: 8, letterSpacing: "0.02em" }}>
          {data.churchName}
        </div>

        <div style={{ marginTop: 24, fontSize: 14, fontFamily: v2Styles.serif, color: v2Styles.ink, lineHeight: 1.55, maxWidth: 360 }}>
          “{data.verseQuote.text}”
          <div style={{ fontSize: 10.5, marginTop: 6, opacity: 0.55, fontFamily: '"Inter", sans-serif', letterSpacing: "0.15em" }}>
            {data.verseQuote.ref.toUpperCase()}
          </div>
        </div>
      </div>
    </div>
  );
}

function V2Back({ data }) {
  return (
    <div style={{ ...v2Styles.half, ...v2Styles.divider }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
        <div style={{ fontSize: 10, letterSpacing: "0.3em", color: v2Styles.navy, textTransform: "uppercase" }}>
          Notice
        </div>
        <div style={{ fontFamily: v2Styles.serif, fontSize: 14, color: v2Styles.navy, opacity: 0.6 }}>
          교회 소식
        </div>
      </div>
      <V2Hairline />

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "0", marginTop: 0, fontSize: 11, lineHeight: 1.6 }}>
        {data.announcements.map((a, i) => (
          <div key={i} style={{
            padding: "16px 20px 16px 0",
            borderBottom: `1px solid ${v2Styles.rule}`,
            borderRight: i % 2 === 0 ? `1px solid ${v2Styles.rule}` : "none",
            paddingLeft: i % 2 === 1 ? 20 : 0,
          }}>
            <div style={{ display: "flex", alignItems: "baseline", gap: 10 }}>
              <span style={{ fontSize: 10, color: v2Styles.navy, opacity: 0.4, fontVariantNumeric: "tabular-nums" }}>
                {String(i + 1).padStart(2, "0")}
              </span>
              <span style={{ fontFamily: v2Styles.serif, fontSize: 14, color: v2Styles.navy }}>
                {a.title}
              </span>
            </div>
            <div style={{ marginTop: 6, color: v2Styles.ink, opacity: 0.75 }}>{a.body}</div>
          </div>
        ))}
      </div>

      <div style={{ marginTop: "auto", display: "flex", justifyContent: "space-between", alignItems: "flex-end" }}>
        <div style={{ fontFamily: v2Styles.serif, fontSize: 28, color: v2Styles.navy, letterSpacing: "0.02em" }}>
          {data.tagline || "오직 은혜."}
        </div>
        <div style={{ fontSize: 10, color: v2Styles.ink, opacity: 0.55, letterSpacing: "0.15em", textAlign: "right" }}>
          {data.website || ""}<br />
          {data.blogInfo || ""}
        </div>
      </div>
    </div>
  );
}

function V2Inside({ data }) {
  return (
    <div style={v2Styles.paper}>
      <div style={v2Styles.half}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
          <div style={{ fontSize: 10, letterSpacing: "0.3em", color: v2Styles.navy, textTransform: "uppercase" }}>
            Order of Worship
          </div>
          <div style={{ fontSize: 10, letterSpacing: "0.2em", color: v2Styles.navy, opacity: 0.5 }}>
            01 / 03
          </div>
        </div>
        <V2Hairline />

        <div style={{ display: "flex", alignItems: "baseline", gap: 14, marginTop: 18 }}>
          <div style={{ fontFamily: v2Styles.serif, fontSize: 40, color: v2Styles.navy, letterSpacing: "0.02em" }}>
            오전 예배
          </div>
          <div style={{ fontSize: 11, letterSpacing: "0.25em", color: v2Styles.ink, opacity: 0.55 }}>
            MORNING SERVICE
          </div>
        </div>
        <div style={{ fontSize: 11.5, color: v2Styles.ink, opacity: 0.65, marginTop: 6 }}>
          본문 · {data.scripture} &nbsp;·&nbsp; 설교 · {data.sermonTitle}
        </div>

        <div style={{ marginTop: 22, fontSize: 12, lineHeight: 2 }}>
          {data.order.filter(o => !o.kind).map((o, i) => (
            <div key={i} style={{
              display: "grid",
              gridTemplateColumns: "32px 88px 1fr 110px",
              gap: 12,
              alignItems: "baseline",
              padding: "4px 0",
              borderBottom: i === data.order.filter(x => !x.kind).length - 1 ? "none" : `0.5px solid ${v2Styles.rule}`,
            }}>
              <div style={{ fontSize: 10, color: v2Styles.navy, opacity: 0.35, fontVariantNumeric: "tabular-nums", letterSpacing: "0.05em" }}>
                {String(i + 1).padStart(2, "0")}
              </div>
              <div style={{ fontFamily: v2Styles.serif, fontSize: 13, color: v2Styles.navy }}>
                {o.item}
              </div>
              <div style={{
                color: v2Styles.ink,
                fontWeight: o.emphasize ? 500 : 400,
                fontSize: o.emphasize ? 13 : 11.5,
                fontFamily: o.emphasize ? v2Styles.serif : "inherit",
              }}>
                {o.content || "—"}
              </div>
              <div style={{ color: v2Styles.ink, opacity: 0.55, fontSize: 10.5, textAlign: "right" }}>
                {o.who}
              </div>
            </div>
          ))}
        </div>

        <div style={{ marginTop: 16, display: "flex", gap: 18, fontSize: 10.5, color: v2Styles.ink, opacity: 0.65 }}>
          {data.helpers.map((h, i) => (
            <span key={i}>
              <span style={{ opacity: 0.5 }}>{h.role}</span>&nbsp;&nbsp;{h.name}
            </span>
          ))}
        </div>
      </div>

      <div style={{ ...v2Styles.half, ...v2Styles.divider }}>
        <div style={{ fontSize: 10, letterSpacing: "0.3em", color: v2Styles.navy, textTransform: "uppercase" }}>
          Today's Word
        </div>
        <V2Hairline />

        {/* 큰 말씀 인용 */}
        <div style={{ marginTop: 32 }}>
          <div style={{ fontFamily: v2Styles.serif, fontSize: 60, lineHeight: 0.9, color: v2Styles.navy, opacity: 0.15 }}>
            “
          </div>
          <div style={{ fontFamily: v2Styles.serif, fontSize: 26, lineHeight: 1.55, color: v2Styles.navy, marginTop: -14, letterSpacing: "0.01em" }}>
            {data.verseQuote.text}
          </div>
          <div style={{ marginTop: 18, fontSize: 11, letterSpacing: "0.3em", color: v2Styles.navy, opacity: 0.6 }}>
            — {data.verseQuote.ref.toUpperCase()}
          </div>
        </div>

        <div style={{ marginTop: 36 }}>
          <V2Hairline />
        </div>

        {/* 오후 / 수요 — 미니 그리드 */}
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 24, marginTop: 22 }}>
          {[
            { title: "오후 예배", en: "AFTERNOON", items: data.afternoon },
            { title: "수요 예배", en: "MIDWEEK", items: data.wednesday },
          ].map((sec, idx) => (
            <div key={idx}>
              <div style={{ display: "flex", alignItems: "baseline", gap: 8 }}>
                <div style={{ fontFamily: v2Styles.serif, fontSize: 18, color: v2Styles.navy }}>{sec.title}</div>
                <div style={{ fontSize: 9, letterSpacing: "0.25em", color: v2Styles.ink, opacity: 0.5 }}>{sec.en}</div>
              </div>
              <div style={{ marginTop: 10, fontSize: 11.5, lineHeight: 1.95 }}>
                {sec.items.map((o, i) => (
                  <div key={i} style={{ display: "flex", justifyContent: "space-between", borderBottom: `0.5px solid ${v2Styles.rule}` }}>
                    <span style={{ fontFamily: v2Styles.serif, color: v2Styles.navy }}>{o.item}</span>
                    <span style={{ color: v2Styles.ink, opacity: 0.7 }}>{o.content || "—"}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div style={{ marginTop: "auto", display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
          <div style={{ fontFamily: v2Styles.serif, fontSize: 11, color: v2Styles.navy, opacity: 0.55, letterSpacing: "0.2em" }}>
            {data.churchName}
          </div>
          <div style={{ fontSize: 10, color: v2Styles.ink, opacity: 0.5, letterSpacing: "0.15em" }}>
            {data.date}
          </div>
        </div>
      </div>
    </div>
  );
}

function V2Outside({ data }) {
  return (
    <div style={v2Styles.paper}>
      <V2Back data={data} />
      <V2Cover data={data} />
    </div>
  );
}

window.V2Outside = V2Outside;
window.V2Inside = V2Inside;
