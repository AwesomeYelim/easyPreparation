// 시안 5 — 청년부 감각
// 대담한 타입 + 그래픽 비율, 베이지 + 짙은 잉크 + 비비드 포인트(따뜻한 오렌지)
// 모던 그래픽, 큰 글자 블록, 단단한 그리드

const v5Styles = {
  paper: {
    width: 1200,
    height: 848,
    background: "#EDE6D7",
    color: "#161412",
    fontFamily: '"Inter", "Noto Sans KR", system-ui, sans-serif',
    position: "relative",
    overflow: "hidden",
    display: "flex",
  },
  half: { flex: 1, padding: 0, position: "relative", display: "flex", flexDirection: "column" },
  ink: "#161412",
  cream: "#EDE6D7",
  point: "#D85B2A",
  pointDk: "#A53F18",
  rule: "#161412",
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

function V5Cover({ data }) {
  return (
    <div style={{ ...v5Styles.half, padding: 0 }}>
      {/* 헤더 바 */}
      <div style={{
        background: v5Styles.ink, color: v5Styles.cream,
        padding: "14px 36px", display: "flex", justifyContent: "space-between", alignItems: "center",
        fontSize: 10, letterSpacing: "0.4em", textTransform: "uppercase",
      }}>
        <span>{`WEEKLY · ${weekNum(data.weekNumber)}`}</span>
        <span style={{ color: v5Styles.point }}>● LIVE WORSHIP</span>
        <span>{`VOL. ${parseDate(data.date).day}.${String(parseDate(data.date).month).padStart(2,'0')}`}</span>
      </div>

      {/* 거대 타이틀 */}
      <div style={{ padding: "44px 36px 24px", flex: 1, display: "flex", flexDirection: "column", justifyContent: "space-between" }}>
        <div>
          <div style={{ fontSize: 11, letterSpacing: "0.5em", color: v5Styles.point, fontWeight: 700 }}>
            {`SUNDAY / ${String(parseDate(data.date).day).padStart(2,'0')}.${String(parseDate(data.date).month).padStart(2,'0')}.${parseDate(data.date).year}`}
          </div>
          <div style={{
            fontSize: 130, lineHeight: 0.9, fontWeight: 900,
            color: v5Styles.ink, letterSpacing: "-0.05em",
            marginTop: 18, fontFamily: v5Styles.serif,
          }}>
            아들이<br/>
            <span style={{ color: v5Styles.point, fontStyle: "italic" }}>자유케</span><br/>
            하면.
          </div>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 24, alignItems: "end" }}>
          <div>
            <div style={{ fontSize: 10, letterSpacing: "0.3em", color: v5Styles.ink, opacity: 0.6 }}>SCRIPTURE</div>
            <div style={{ fontFamily: v5Styles.serif, fontSize: 22, color: v5Styles.ink, marginTop: 6 }}>
              {data.scripture}
            </div>
          </div>
          <div>
            <div style={{ fontSize: 10, letterSpacing: "0.3em", color: v5Styles.ink, opacity: 0.6 }}>PREACHED BY</div>
            <div style={{ fontFamily: v5Styles.serif, fontSize: 22, color: v5Styles.ink, marginTop: 6 }}>
              {data.pastor}
            </div>
          </div>
        </div>
      </div>

      {/* 풋터 */}
      <div style={{
        background: v5Styles.point, color: v5Styles.cream,
        padding: "18px 36px",
      }}>
        <div style={{ fontSize: 10, letterSpacing: "0.3em", opacity: 0.85 }}>{data.churchNameEn}</div>
        <div style={{ fontSize: 28, fontWeight: 800, marginTop: 4, letterSpacing: "-0.01em" }}>
          {data.churchName}
        </div>
      </div>
    </div>
  );
}

function V5Back({ data }) {
  return (
    <div style={{ ...v5Styles.half, borderLeft: `1.5px solid ${v5Styles.ink}` }}>
      <div style={{
        background: v5Styles.cream, color: v5Styles.ink,
        padding: "14px 36px", display: "flex", justifyContent: "space-between", alignItems: "center",
        fontSize: 10, letterSpacing: "0.4em", textTransform: "uppercase",
        borderBottom: `1.5px solid ${v5Styles.ink}`,
      }}>
        <span>NOTICE BOARD</span>
        <span style={{ color: v5Styles.point }}>● UPDATED</span>
      </div>

      <div style={{ padding: "24px 36px", flex: 1, display: "flex", flexDirection: "column", gap: 0 }}>
        {data.announcements.map((a, i) => (
          <div key={i} style={{
            display: "grid",
            gridTemplateColumns: "60px 1fr",
            gap: 18,
            padding: "12px 0",
            borderBottom: i === data.announcements.length - 1 ? "none" : `1px solid ${v5Styles.ink}`,
          }}>
            <div style={{
              fontFamily: v5Styles.serif, fontSize: 36, lineHeight: 0.9,
              color: i % 3 === 0 ? v5Styles.point : v5Styles.ink,
              fontWeight: 800, letterSpacing: "-0.04em",
            }}>
              {String(i + 1).padStart(2, "0")}
            </div>
            <div>
              <div style={{ fontSize: 14, fontWeight: 700, color: v5Styles.ink, letterSpacing: "0.01em" }}>
                {a.title}
              </div>
              <div style={{ fontSize: 11, color: v5Styles.ink, opacity: 0.78, lineHeight: 1.55, marginTop: 3 }}>
                {a.body}
              </div>
            </div>
          </div>
        ))}
      </div>

      <div style={{
        background: v5Styles.ink, color: v5Styles.cream,
        padding: "16px 36px", display: "flex", justifyContent: "space-between",
        fontSize: 10, letterSpacing: "0.25em",
      }}>
        <span>{data.website || ""}</span>
        <span style={{ color: v5Styles.point }}>{(data.blogInfo || "").toUpperCase()}</span>
      </div>
    </div>
  );
}

function V5Inside({ data }) {
  return (
    <div style={v5Styles.paper}>
      <div style={v5Styles.half}>
        <div style={{
          background: v5Styles.cream, color: v5Styles.ink,
          padding: "14px 36px", display: "flex", justifyContent: "space-between", alignItems: "center",
          fontSize: 10, letterSpacing: "0.4em", borderBottom: `1.5px solid ${v5Styles.ink}`,
        }}>
          <span>ORDER OF WORSHIP</span>
          <span>SUN · MORNING</span>
        </div>

        <div style={{ padding: "30px 36px 0" }}>
          <div style={{
            fontFamily: v5Styles.serif, fontSize: 96, lineHeight: 0.9,
            fontWeight: 900, letterSpacing: "-0.04em", color: v5Styles.ink,
          }}>
            예배순서<span style={{ color: v5Styles.point }}>.</span>
          </div>
          <div style={{ fontSize: 11, letterSpacing: "0.25em", color: v5Styles.ink, opacity: 0.7, marginTop: 8 }}>
            01 — 19 · MORNING SERVICE
          </div>
        </div>

        <div style={{ padding: "20px 36px 0", fontSize: 12, lineHeight: 1.85 }}>
          {data.order.filter(o => !o.kind).map((o, i) => (
            <div key={i} style={{
              display: "grid",
              gridTemplateColumns: "30px 90px 1fr 90px",
              gap: 10,
              alignItems: "baseline",
              padding: "4px 0",
              borderBottom: `1px solid rgba(22,20,18,0.15)`,
              background: o.emphasize ? "rgba(216,91,42,0.1)" : "transparent",
              paddingLeft: o.emphasize ? 8 : 0,
              borderLeft: o.emphasize ? `3px solid ${v5Styles.point}` : "3px solid transparent",
            }}>
              <span style={{ fontSize: 10, fontWeight: 700, color: v5Styles.point, fontVariantNumeric: "tabular-nums", letterSpacing: "0.05em" }}>
                {String(i + 1).padStart(2, "0")}
              </span>
              <span style={{ fontFamily: v5Styles.serif, fontSize: 13, fontWeight: 700, color: v5Styles.ink }}>
                {o.item}
              </span>
              <span style={{
                color: v5Styles.ink, opacity: 0.85,
                fontWeight: o.emphasize ? 700 : 400,
                fontSize: o.emphasize ? 13.5 : 11.5,
                fontFamily: o.emphasize ? v5Styles.serif : "inherit",
              }}>
                {o.content || "—"}
              </span>
              <span style={{ fontSize: 10, color: v5Styles.ink, opacity: 0.6, textAlign: "right", letterSpacing: "0.05em" }}>
                {o.who}
              </span>
            </div>
          ))}
        </div>

        <div style={{ padding: "12px 36px", marginTop: "auto", fontSize: 10.5, color: v5Styles.ink, opacity: 0.7, letterSpacing: "0.05em" }}>
          {data.helpers.map(h => `${h.role} ${h.name}`).join("   /   ")}
        </div>
      </div>

      <div style={{ ...v5Styles.half, borderLeft: `1.5px solid ${v5Styles.ink}` }}>
        {/* 풀블리드 말씀 — 포인트 컬러 블록 */}
        <div style={{ background: v5Styles.ink, color: v5Styles.cream, padding: "32px 36px" }}>
          <div style={{ fontSize: 10, letterSpacing: "0.4em", color: v5Styles.point }}>
            ▲ TODAY'S WORD
          </div>
          <div style={{
            fontFamily: v5Styles.serif, fontSize: 30, lineHeight: 1.45, fontWeight: 600,
            marginTop: 16, letterSpacing: "0.01em",
          }}>
            “{data.verseQuote.text}”
          </div>
          <div style={{ fontSize: 11, letterSpacing: "0.3em", marginTop: 16, color: v5Styles.point }}>
            — {data.verseQuote.ref.toUpperCase()}
          </div>
        </div>

        {/* 오후 / 수요 */}
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", flex: 1 }}>
          {[
            { title: "오후 예배", en: "AFTERNOON", items: data.afternoon, num: "02" },
            { title: "수요 예배", en: "WEDNESDAY", items: data.wednesday, num: "03" },
          ].map((sec, idx) => (
            <div key={idx} style={{
              padding: "26px 28px",
              borderRight: idx === 0 ? `1.5px solid ${v5Styles.ink}` : "none",
            }}>
              <div style={{ display: "flex", alignItems: "baseline", gap: 8 }}>
                <span style={{ fontFamily: v5Styles.serif, fontSize: 32, fontWeight: 900, color: v5Styles.point, lineHeight: 1, letterSpacing: "-0.04em" }}>
                  {sec.num}
                </span>
                <span style={{ fontFamily: v5Styles.serif, fontSize: 22, fontWeight: 700, color: v5Styles.ink }}>
                  {sec.title}
                </span>
              </div>
              <div style={{ fontSize: 9, letterSpacing: "0.3em", color: v5Styles.ink, opacity: 0.6, marginTop: 4 }}>
                {sec.en} SERVICE
              </div>
              <div style={{ marginTop: 16, fontSize: 12, lineHeight: 2.05 }}>
                {sec.items.map((o, j) => (
                  <div key={j} style={{ display: "flex", justifyContent: "space-between", borderBottom: `1px solid rgba(22,20,18,0.15)` }}>
                    <span style={{ fontFamily: v5Styles.serif, fontWeight: 700, color: v5Styles.ink }}>{o.item}</span>
                    <span style={{ color: v5Styles.ink, opacity: 0.75 }}>{o.content || "—"}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div style={{
          background: v5Styles.point, color: v5Styles.cream,
          padding: "16px 36px", display: "flex", justifyContent: "space-between",
          fontSize: 10, letterSpacing: "0.3em", fontWeight: 700,
        }}>
          <span>{data.churchName.toUpperCase()}</span>
          <span>{data.date}</span>
        </div>
      </div>
    </div>
  );
}

function V5Outside({ data }) {
  return (
    <div style={v5Styles.paper}>
      <V5Back data={data} />
      <V5Cover data={data} />
    </div>
  );
}

window.V5Outside = V5Outside;
window.V5Inside = V5Inside;
