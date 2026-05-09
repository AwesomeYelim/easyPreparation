// 시안 6 — 정갈한 한국형 클래식
// 따뜻한 베이지 + 깊은 네이비, 명조 중심, 한자/영문 부제 없이 한글만으로 격조

const v6Styles = {
  paper: {
    width: 1200,
    height: 848,
    background: "#F1ECDF",
    color: "#1A2336",
    fontFamily: '"Noto Serif KR", "Nanum Myeongjo", serif',
    position: "relative",
    overflow: "hidden",
    display: "flex",
  },
  half: { flex: 1, padding: "60px 56px", position: "relative", display: "flex", flexDirection: "column" },
  divider: { borderLeft: "1px solid #C9BFA4" },
  navy: "#1A2336",
  navyLt: "#3A4760",
  bg: "#F1ECDF",
  rule: "#C9BFA4",
  sans: '"Noto Sans KR", system-ui, sans-serif',
};

function V6Rule({ thin = false }) {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 3, margin: "10px 0" }}>
      <div style={{ height: thin ? 0.5 : 1, background: v6Styles.navy }} />
      <div style={{ height: thin ? 0.5 : 0.5, background: v6Styles.navy, opacity: thin ? 0 : 0.4 }} />
    </div>
  );
}

function V6Cover({ data }) {
  return (
    <div style={{ ...v6Styles.half, justifyContent: "space-between", alignItems: "center", textAlign: "center" }}>
      <div style={{ width: "100%" }}>
        <div style={{ fontFamily: v6Styles.sans, fontSize: 11, letterSpacing: "0.45em", color: v6Styles.navy, fontWeight: 500 }}>
          주 일 예 배 순 서
        </div>
        <V6Rule />
      </div>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", alignItems: "center", gap: 36 }}>
        {/* 표지 이미지 or 모노그램 */}
        {data.coverImage ? (
          <div style={{
            width: 160, height: 160, borderRadius: "50%",
            border: `1px solid ${v6Styles.navy}`,
            overflow: "hidden",
            backgroundImage: `url(${data.coverImage})`,
            backgroundSize: "cover",
            backgroundPosition: "center",
          }} />
        ) : (
          <div style={{
            width: 130, height: 130, borderRadius: "50%",
            border: `1px solid ${v6Styles.navy}`,
            display: "flex", alignItems: "center", justifyContent: "center",
            position: "relative",
          }}>
            <div style={{ position: "absolute", inset: 9, borderRadius: "50%", border: `0.5px solid ${v6Styles.navy}`, opacity: 0.4 }} />
            <div style={{ fontSize: 44, color: v6Styles.navy, lineHeight: 1, letterSpacing: 0 }}>✠</div>
          </div>
        )}

        <div>
          <div style={{ fontSize: 46, lineHeight: 1.25, color: v6Styles.navy, letterSpacing: "0.04em" }}>
            {data.churchName}
          </div>
          <div style={{ fontFamily: v6Styles.sans, fontSize: 12, marginTop: 18, color: v6Styles.navyLt, letterSpacing: "0.5em" }}>
            오 직 은 혜
          </div>
        </div>

        <div style={{ maxWidth: 400, lineHeight: 1.6, color: v6Styles.navy, wordBreak: "keep-all" }}>
          <div style={{ fontSize: 19, fontStyle: "italic" }}>
            “{data.verseQuote.text}”
          </div>
          <div style={{ fontFamily: v6Styles.sans, fontSize: 11, marginTop: 12, color: v6Styles.navyLt, letterSpacing: "0.2em" }}>
            {data.verseQuote.ref}
          </div>
        </div>
      </div>

      <div style={{ width: "100%" }}>
        <V6Rule />
        <div style={{ fontFamily: v6Styles.sans, display: "flex", justifyContent: "space-between", fontSize: 11, letterSpacing: "0.15em", color: v6Styles.navy }}>
          <span>{data.date}</span>
          <span>{data.weekNumber}</span>
          <span>{data.pastor}</span>
        </div>
      </div>
    </div>
  );
}

function V6Back({ data }) {
  return (
    <div style={{ ...v6Styles.half, ...v6Styles.divider }}>
      <div style={{ fontFamily: v6Styles.sans, fontSize: 11, letterSpacing: "0.45em", color: v6Styles.navy, fontWeight: 500 }}>
        교 회 소 식
      </div>
      <V6Rule />

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px 32px", marginTop: 14, fontFamily: v6Styles.sans, fontSize: 11, lineHeight: 1.65 }}>
        {data.announcements.map((a, i) => (
          <div key={i}>
            <div style={{ fontFamily: '"Noto Serif KR", serif', fontSize: 14, color: v6Styles.navy, marginBottom: 5, letterSpacing: "0.04em" }}>
              <span style={{ fontFamily: v6Styles.sans, fontSize: 10, color: v6Styles.navyLt, marginRight: 8, fontVariantNumeric: "tabular-nums" }}>
                {String(i + 1).padStart(2, "0")}
              </span>
              {a.title}
            </div>
            <div style={{ color: v6Styles.navy, opacity: 0.82, wordBreak: "keep-all" }}>{a.body}</div>
          </div>
        ))}
      </div>

      <div style={{ marginTop: "auto" }}>
        <V6Rule thin />
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", fontFamily: v6Styles.sans, fontSize: 10, color: v6Styles.navyLt, letterSpacing: "0.1em" }}>
          <div>
            <div style={{ fontFamily: '"Noto Serif KR", serif', fontSize: 17, color: v6Styles.navy, letterSpacing: "0.1em" }}>
              {data.tagline || "오직 하나님께만 영광을"}
            </div>
          </div>
          <div style={{ textAlign: "right" }}>
            <div>홈페이지 · {data.website || ""}</div>
            <div style={{ marginTop: 3 }}>블로그 · {data.blogInfo || ""}</div>
          </div>
        </div>
      </div>
    </div>
  );
}

function V6Inside({ data }) {
  return (
    <div style={v6Styles.paper}>
      <div style={v6Styles.half}>
        <div style={{ fontFamily: v6Styles.sans, fontSize: 11, letterSpacing: "0.45em", color: v6Styles.navy, fontWeight: 500 }}>
          오 전 예 배
        </div>
        <V6Rule />
        <div style={{ fontFamily: v6Styles.sans, fontSize: 11, color: v6Styles.navyLt, letterSpacing: "0.1em", marginTop: 4 }}>
          본문 · {data.scripture}　／　설교 · {data.sermonTitle}
        </div>

        <div style={{ marginTop: 20, fontSize: 12.5, lineHeight: 2 }}>
          {data.order.filter(o => !o.kind).map((o, i) => (
            <div key={i} style={{
              display: "grid",
              gridTemplateColumns: "100px 1fr 130px",
              gap: 14,
              alignItems: "baseline",
              borderBottom: `0.5px dotted ${v6Styles.rule}`,
              padding: "5px 0",
            }}>
              <div style={{ color: v6Styles.navy, letterSpacing: "0.18em", fontSize: 12.5 }}>
                {o.item}
              </div>
              <div style={{
                fontFamily: v6Styles.sans,
                color: v6Styles.navy,
                opacity: 0.85,
                fontSize: o.emphasize ? 13 : 11.5,
                fontStyle: o.emphasize ? "italic" : "normal",
                fontWeight: o.emphasize ? 500 : 400,
              }}>
                {o.content || "—"}
              </div>
              <div style={{ fontFamily: v6Styles.sans, color: v6Styles.navyLt, fontSize: 10.5, textAlign: "right", letterSpacing: "0.1em" }}>
                {o.who}
              </div>
            </div>
          ))}
        </div>

        <div style={{ marginTop: 18, fontFamily: v6Styles.sans, fontSize: 10.5, color: v6Styles.navyLt, letterSpacing: "0.1em" }}>
          {data.helpers.map(h => `${h.role} ─ ${h.name}`).join("　·　")}
        </div>
      </div>

      <div style={{ ...v6Styles.half, ...v6Styles.divider }}>
        <div style={{ fontFamily: v6Styles.sans, fontSize: 11, letterSpacing: "0.45em", color: v6Styles.navy, fontWeight: 500 }}>
          오늘의 말씀
        </div>
        <V6Rule />

        {/* 말씀 강조 */}
        <div style={{
          marginTop: 12,
          padding: "30px 30px",
          border: `1px solid ${v6Styles.navy}`,
          background: "rgba(26,35,54,0.04)",
          position: "relative",
        }}>
          <div style={{ fontSize: 22, lineHeight: 1.65, color: v6Styles.navy, letterSpacing: "0.03em", wordBreak: "keep-all" }}>
            “{data.verseQuote.text}”
          </div>
          <div style={{ fontFamily: v6Styles.sans, fontSize: 11, marginTop: 16, color: v6Styles.navyLt, letterSpacing: "0.2em", textAlign: "right" }}>
            — {data.verseQuote.ref}
          </div>
        </div>

        <div style={{ marginTop: 28 }}>
          <div style={{ fontFamily: v6Styles.sans, fontSize: 11, letterSpacing: "0.45em", color: v6Styles.navy, fontWeight: 500 }}>
            오 후 · 수 요 예 배
          </div>
          <V6Rule thin />
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 32, marginTop: 6 }}>
            {[
              { title: "오후 예배", items: data.afternoon },
              { title: "수요 예배", items: data.wednesday },
            ].map((sec, i) => (
              <div key={i}>
                <div style={{ fontSize: 17, color: v6Styles.navy, letterSpacing: "0.1em", marginBottom: 8 }}>{sec.title}</div>
                <div style={{ fontFamily: v6Styles.sans, fontSize: 11.5, lineHeight: 2.1 }}>
                  {sec.items.map((o, j) => (
                    <div key={j} style={{ display: "flex", justifyContent: "space-between", borderBottom: `0.5px dotted ${v6Styles.rule}` }}>
                      <span style={{ fontFamily: '"Noto Serif KR", serif', color: v6Styles.navy, letterSpacing: "0.1em" }}>{o.item}</span>
                      <span style={{ color: v6Styles.navy, opacity: 0.78 }}>{o.content || "—"}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div style={{ marginTop: "auto" }}>
          <V6Rule thin />
          <div style={{ fontFamily: v6Styles.sans, fontSize: 10, color: v6Styles.navyLt, textAlign: "center", letterSpacing: "0.3em" }}>
            {data.churchName}　·　{data.date}
          </div>
        </div>
      </div>
    </div>
  );
}

function V6Outside({ data }) {
  return (
    <div style={v6Styles.paper}>
      <V6Back data={data} />
      <V6Cover data={data} />
    </div>
  );
}

window.V6Outside = V6Outside;
window.V6Inside = V6Inside;
