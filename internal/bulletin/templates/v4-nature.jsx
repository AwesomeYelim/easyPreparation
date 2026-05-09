// 시안 4 — 자연/은혜 (식물·빛 모티브)
// 올리브 그린 + 따뜻한 베이지/크림, 부드러운 곡선, 잎사귀 placeholder

const v4Styles = {
  paper: {
    width: 1200,
    height: 848,
    background: "#F4F0E4",
    color: "#23332A",
    fontFamily: '"Noto Sans KR", "Inter", system-ui, sans-serif',
    position: "relative",
    overflow: "hidden",
    display: "flex",
    boxSizing: "border-box",
  },
  half: { flex: 1, minWidth: 0, padding: "56px 52px", position: "relative", display: "flex", flexDirection: "column", boxSizing: "border-box" },
  ink: "#23332A",
  cream: "#F4F0E4",
  olive: (typeof window !== 'undefined' && window.__BULLETIN_ACCENT__) || "#5A6A3D",
  oliveDk: "#3F4A2A",
  rule: "#C7BFA6",
  serif: '"Noto Serif KR", "Nanum Myeongjo", serif',
};

// 단순 잎사귀 placeholder — 줄무늬 SVG와 모노스페이스 라벨
function V4LeafPlaceholder({ size = 140, label = "leaf illustration" }) {
  return (
    <div style={{ position: "relative", width: size, height: size }}>
      <svg viewBox="0 0 100 100" width={size} height={size} style={{ display: "block" }}>
        <defs>
          <pattern id="v4stripes" patternUnits="userSpaceOnUse" width="6" height="6" patternTransform="rotate(45)">
            <line x1="0" y1="0" x2="0" y2="6" stroke={v4Styles.olive} strokeWidth="1" opacity="0.35" />
          </pattern>
        </defs>
        <ellipse cx="50" cy="50" rx="38" ry="46" fill="url(#v4stripes)" stroke={v4Styles.olive} strokeWidth="0.8" />
        <line x1="50" y1="6" x2="50" y2="94" stroke={v4Styles.olive} strokeWidth="0.6" opacity="0.5" />
      </svg>
      <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: "ui-monospace, monospace", fontSize: 9, color: v4Styles.oliveDk, letterSpacing: "0.1em" }}>
        {label}
      </div>
    </div>
  );
}

function V4Cover({ data }) {
  return (
    <div style={{ ...v4Styles.half, alignItems: "center", textAlign: "center", justifyContent: "space-between" }}>
      <div style={{ width: "100%", display: "flex", justifyContent: "space-between", fontSize: 10, letterSpacing: "0.3em", color: v4Styles.olive }}>
        <span>· GRACE ·</span>
        <span>· LIGHT ·</span>
        <span>· LIFE ·</span>
      </div>

      {data.coverImage ? (
        <div style={{
          width: 280, height: 220,
          backgroundImage: `url(${data.coverImage})`,
          backgroundSize: "cover", backgroundPosition: "center",
          borderRadius: 12,
          border: `1px solid ${v4Styles.olive}`,
        }} />
      ) : (
        <V4LeafPlaceholder size={180} label="leaf · 잎사귀" />
      )}

      <div>
        <div style={{ fontFamily: v4Styles.serif, fontSize: 13, color: v4Styles.olive, letterSpacing: "0.5em" }}>
          주 일  예 배
        </div>
        <div style={{ fontFamily: v4Styles.serif, fontSize: 46, color: v4Styles.oliveDk, marginTop: 14, letterSpacing: "0.02em", lineHeight: 1.2 }}>
          {data.churchName}
        </div>
        <div style={{ fontSize: 10, letterSpacing: "0.4em", color: v4Styles.olive, opacity: 0.8, marginTop: 12 }}>
          {data.churchNameEn}
        </div>
      </div>

      <div style={{ maxWidth: 380, fontFamily: v4Styles.serif, fontSize: 16, lineHeight: 1.6, color: v4Styles.ink, fontStyle: "italic", wordBreak: "keep-all" }}>
        “{data.verseQuote.text}”
        <div style={{ fontSize: 11, marginTop: 10, color: v4Styles.olive, fontStyle: "normal", letterSpacing: "0.2em" }}>
          {data.verseQuote.ref}
        </div>
      </div>

      <div style={{ width: "100%", paddingTop: 18, borderTop: `1px solid ${v4Styles.rule}`, display: "flex", justifyContent: "space-between", fontSize: 11, color: v4Styles.oliveDk, letterSpacing: "0.15em" }}>
        <span>{data.date}</span>
        <span>{data.weekNumber}</span>
        <span>{data.pastor}</span>
      </div>
    </div>
  );
}

function V4Back({ data }) {
  return (
    <div style={{ ...v4Styles.half, borderLeft: `1px solid ${v4Styles.rule}` }}>
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        <V4LeafPlaceholder size={32} label="" />
        <div>
          <div style={{ fontFamily: v4Styles.serif, fontSize: 26, color: v4Styles.oliveDk, letterSpacing: "0.05em" }}>
            교회 소식
          </div>
          <div style={{ fontSize: 9, letterSpacing: "0.4em", color: v4Styles.olive, marginTop: 2 }}>
            ANNOUNCEMENTS
          </div>
        </div>
      </div>

      <div style={{ marginTop: 24, display: "flex", flexDirection: "column", gap: data.announcements.length > 10 ? 6 : 14, fontSize: data.announcements.length > 10 ? 9.5 : data.announcements.length > 7 ? 10.5 : 11.5, lineHeight: data.announcements.length > 10 ? 1.4 : 1.6 }}>
        {data.announcements.map((a, i) => (
          <div key={i} style={{
            display: "grid",
            gridTemplateColumns: "auto 1fr",
            gap: data.announcements.length > 10 ? 10 : 16,
            paddingBottom: data.announcements.length > 10 ? 6 : 12,
            borderBottom: i === data.announcements.length - 1 ? "none" : `0.5px dashed ${v4Styles.rule}`,
          }}>
            <div style={{
              fontFamily: v4Styles.serif, fontSize: 13, color: v4Styles.cream,
              background: v4Styles.olive, padding: "6px 12px",
              alignSelf: "flex-start", borderRadius: 999,
              letterSpacing: "0.05em", whiteSpace: "nowrap",
            }}>
              {a.title}
            </div>
            <div style={{ color: v4Styles.ink, opacity: 0.85, paddingTop: 6, wordBreak: "keep-all" }}>{a.body}</div>
          </div>
        ))}
      </div>

      <div style={{ marginTop: "auto", display: "flex", justifyContent: "space-between", alignItems: "flex-end" }}>
        <div style={{ fontFamily: v4Styles.serif, fontStyle: "italic", fontSize: 22, color: v4Styles.olive }}>
          {data.tagline || "심으신 자리에서 꽃피우는 신앙"}
        </div>
        <div style={{ fontSize: 10, color: v4Styles.oliveDk, textAlign: "right", letterSpacing: "0.15em", opacity: 0.75 }}>
          {data.website || ""}<br/>{data.blogInfo || ""}
        </div>
      </div>
    </div>
  );
}

function V4Inside({ data }) {
  return (
    <div style={v4Styles.paper}>
      <div style={v4Styles.half}>
        <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 4 }}>
          <V4LeafPlaceholder size={32} label="" />
          <div>
            <div style={{ fontFamily: v4Styles.serif, fontSize: 26, color: v4Styles.oliveDk }}>오전 예배</div>
            <div style={{ fontSize: 9, letterSpacing: "0.4em", color: v4Styles.olive }}>MORNING WORSHIP · {data.scripture}</div>
          </div>
        </div>

        <div style={{ marginTop: 22, fontSize: 12, lineHeight: 1.95 }}>
          {data.order.filter(o => !o.kind).map((o, i) => (
            <div key={i} style={{
              display: "grid",
              gridTemplateColumns: "10px 90px 1fr 100px",
              gap: 12,
              alignItems: "baseline",
              padding: "5px 0",
              borderBottom: `0.5px dashed ${v4Styles.rule}`,
            }}>
              <span style={{ width: 6, height: 6, borderRadius: "50%", background: o.emphasize ? v4Styles.olive : "transparent", border: `1px solid ${v4Styles.olive}`, display: "block", alignSelf: "center" }} />
              <span style={{ fontFamily: v4Styles.serif, fontSize: 13, color: v4Styles.oliveDk }}>{o.item}</span>
              <span style={{ color: v4Styles.ink, opacity: 0.85, fontStyle: o.emphasize ? "italic" : "normal", fontSize: o.emphasize ? 13 : 11.5 }}>
                {o.content || "—"}
              </span>
              <span style={{ fontSize: 10, color: v4Styles.olive, textAlign: "right", letterSpacing: "0.05em" }}>{o.who}</span>
            </div>
          ))}
        </div>

        <div style={{ marginTop: 14, padding: "10px 14px", background: "rgba(90,106,61,0.1)", borderRadius: 6, fontSize: 10.5, color: v4Styles.oliveDk, letterSpacing: "0.05em" }}>
          {data.helpers.map(h => `${h.role}  ${h.name}`).join("    ·    ")}
        </div>
      </div>

      <div style={{ ...v4Styles.half, borderLeft: `1px solid ${v4Styles.rule}` }}>
        {/* 큰 말씀 카드 */}
        <div style={{
          background: v4Styles.olive,
          color: v4Styles.cream,
          padding: "30px 30px",
          borderRadius: 12,
          position: "relative",
          overflow: "hidden",
        }}>
          <div style={{ position: "absolute", right: -20, bottom: -20, opacity: 0.2 }}>
            <V4LeafPlaceholder size={150} label="" />
          </div>
          <div style={{ fontSize: 9, letterSpacing: "0.4em", opacity: 0.85 }}>오늘의 말씀</div>
          <div style={{ fontFamily: v4Styles.serif, fontSize: 24, lineHeight: 1.55, marginTop: 12, position: "relative", zIndex: 1, wordBreak: "keep-all" }}>
            “{data.verseQuote.text}”
          </div>
          <div style={{ fontSize: 11, letterSpacing: "0.25em", marginTop: 14, opacity: 0.9 }}>
            — {data.verseQuote.ref}
          </div>
        </div>

        {/* 오후 + 수요 */}
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 20, marginTop: 24 }}>
          {[
            { title: "오후 예배", en: "AFTERNOON", items: data.afternoon },
            { title: "수요 예배", en: "WEDNESDAY", items: data.wednesday },
          ].map((sec, i) => (
            <div key={i} style={{
              padding: "20px 22px",
              border: `1px solid ${v4Styles.rule}`,
              borderRadius: 8,
              background: "rgba(255,255,255,0.4)",
            }}>
              <div style={{ fontFamily: v4Styles.serif, fontSize: 20, color: v4Styles.oliveDk }}>{sec.title}</div>
              <div style={{ fontSize: 9, letterSpacing: "0.3em", color: v4Styles.olive, marginTop: 2 }}>{sec.en}</div>
              <div style={{ marginTop: 12, fontSize: 11.5, lineHeight: 2 }}>
                {sec.items.map((o, j) => (
                  <div key={j} style={{ display: "flex", justifyContent: "space-between", borderBottom: `0.5px dashed ${v4Styles.rule}` }}>
                    <span style={{ fontFamily: v4Styles.serif, color: v4Styles.oliveDk }}>{o.item}</span>
                    <span style={{ color: v4Styles.ink, opacity: 0.75 }}>{o.content || "—"}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div style={{ marginTop: "auto", paddingTop: 14, borderTop: `1px solid ${v4Styles.rule}`, display: "flex", justifyContent: "space-between", fontSize: 10, color: v4Styles.olive, letterSpacing: "0.2em" }}>
          <span>{data.churchName}</span>
          <span>{data.date}</span>
        </div>
      </div>
    </div>
  );
}

function V4Outside({ data }) {
  return (
    <div style={v4Styles.paper}>
      <V4Back data={data} />
      <V4Cover data={data} />
    </div>
  );
}

window.V4Outside = V4Outside;
window.V4Inside = V4Inside;
