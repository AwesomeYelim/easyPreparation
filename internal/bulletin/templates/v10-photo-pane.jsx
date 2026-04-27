// 시안 10 — 사진집 / 무인양품 스타일
// 한 장의 큰 사진 + 명료한 한글 타이포 + 색감 거의 없음 (잉크 한 가지)
// 라운드 모서리 없음, 그리드 대칭 없음, 영어 라벨 최소화

const v10Styles = {
  paper: {
    width: 1200, height: 848,
    background: "#f3f1ec",
    color: "#222",
    display: "flex",
    fontFamily: "'Noto Sans KR', system-ui, sans-serif",
    position: "relative",
    overflow: "hidden",
  },
  ink: "#1f1f1c",
  inkLt: "#7a7670",
  bg: "#f3f1ec",
  bgWhite: "#fbfaf6",
  rule: "#d8d4c9",
  blue: (typeof window !== 'undefined' && window.__BULLETIN_ACCENT__) || "#3a5a78",
  blueLt: "#dee5ed",
  coral: "#c66a4f",
  serif: "'Nanum Myeongjo', 'Noto Serif KR', serif",
  sans: "'Noto Sans KR', system-ui, sans-serif",
};

// 빈 사진 placeholder — 줄친 회색
function V10Photo({ src, height = "100%", caption }) {
  return (
    <div style={{
      width: "100%", height,
      background: src ? "#000" : "#d8d4c9",
      backgroundImage: src ? `url(${src})` : "repeating-linear-gradient(45deg, #d8d4c9 0 12px, #cdc8bb 12px 24px)",
      backgroundSize: "cover", backgroundPosition: "center",
      position: "relative",
    }}>
      {!src && (
        <div style={{
          position: "absolute", inset: 0,
          display: "flex", alignItems: "center", justifyContent: "center",
          color: "#7a7670", fontSize: 11, letterSpacing: "0.4em",
          fontFamily: v10Styles.sans,
        }}>
          ─ 사진 ─
        </div>
      )}
      {caption && (
        <div style={{
          position: "absolute", left: 12, bottom: 10,
          background: "rgba(255,255,255,0.92)",
          color: v10Styles.ink, padding: "5px 10px",
          fontSize: 10, letterSpacing: "0.25em",
        }}>
          {caption}
        </div>
      )}
    </div>
  );
}

function V10Cover({ data }) {
  return (
    <div style={{ flex: 1, display: "flex", flexDirection: "column", borderRight: `1px solid ${v10Styles.rule}`, position: "relative" }}>
      {/* 풀블리드 사진 (상단 ~62%) */}
      <div style={{ height: "62%", position: "relative" }}>
        <V10Photo src={data.coverImage} caption={`${data.date}`} />
      </div>

      {/* 하단 — 단단하고 비대칭한 타입 영역 */}
      <div style={{
        flex: 1, padding: "32px 44px 30px",
        background: v10Styles.bgWhite,
        display: "flex", flexDirection: "column", justifyContent: "space-between",
      }}>
        {/* 한 줄 작은 라벨 */}
        <div style={{
          display: "flex", justifyContent: "space-between",
          fontFamily: v10Styles.sans, fontSize: 10.5,
          letterSpacing: "0.25em", color: v10Styles.inkLt,
          paddingBottom: 8,
        }}>
          <span>{data.weekNumber}</span>
          <span>{data.date}</span>
        </div>

        <div>
          <div style={{
            fontFamily: v10Styles.serif, fontWeight: 700,
            fontSize: 38, letterSpacing: "0.02em", lineHeight: 1.2,
            color: v10Styles.blue, marginBottom: 14,
          }}>
            {data.churchName}
          </div>
          <div style={{
            display: "grid", gridTemplateColumns: "auto 1fr",
            columnGap: 16, rowGap: 4,
            fontSize: 13, fontFamily: v10Styles.sans,
            color: v10Styles.ink,
          }}>
            <span style={{ color: v10Styles.inkLt }}>설교</span>
            <span style={{ fontWeight: 600 }}>{data.sermonTitle}</span>
            <span style={{ color: v10Styles.inkLt }}>본문</span>
            <span>{data.scripture}</span>
            <span style={{ color: v10Styles.inkLt }}>설교자</span>
            <span>{data.pastor}</span>
          </div>
        </div>
      </div>

      {/* 페이지 번호 — 우상단 (청자색) */}
      <div style={{
        position: "absolute", right: 22, top: 22,
        fontFamily: v10Styles.sans, fontSize: 10,
        letterSpacing: "0.3em", color: "#fff",
        background: v10Styles.blue,
        padding: "4px 10px",
      }}>
        ─ 표지 ─
      </div>
    </div>
  );
}

function V10Back({ data }) {
  return (
    <div style={{ flex: 1, padding: "40px 44px", display: "flex", flexDirection: "column", background: v10Styles.bgWhite }}>
      {/* 제목 — 비대칭, 좌측 강조 */}
      <div style={{
        display: "flex", alignItems: "baseline", justifyContent: "space-between",
        marginBottom: 4, paddingBottom: 8,
        borderBottom: `2px solid ${v10Styles.blue}`,
      }}>
        <div style={{
          fontFamily: v10Styles.serif, fontWeight: 700,
          fontSize: 30, letterSpacing: "0.02em",
          color: v10Styles.blue,
        }}>
          교회 소식
        </div>
        <div style={{
          fontFamily: v10Styles.sans, fontSize: 10.5,
          letterSpacing: "0.25em", color: v10Styles.inkLt,
        }}>
          {data.weekNumber}
        </div>
      </div>

      {/* 소식 — 단순한 세로 리스트 (그리드 대칭 없음) */}
      <div style={{ marginTop: 16, fontFamily: v10Styles.sans, flex: 1 }}>
        {data.announcements.map((a, i) => (
          <div key={i} style={{
            display: "grid",
            gridTemplateColumns: "44px 1fr",
            gap: 16,
            padding: "14px 0",
            borderBottom: i === data.announcements.length - 1 ? "none" : `1px solid ${v10Styles.rule}`,
            alignItems: "baseline",
          }}>
            <span style={{
              fontFamily: v10Styles.serif, fontSize: 26,
              fontWeight: 400, color: v10Styles.coral,
              lineHeight: 1, letterSpacing: "-0.02em",
              fontVariantNumeric: "tabular-nums",
            }}>
              {String(i + 1).padStart(2, "0")}
            </span>
            <div>
              <div style={{ fontSize: 15, fontWeight: 700, marginBottom: 3 }}>
                {a.title}
              </div>
              <div style={{ fontSize: 12, lineHeight: 1.65, color: v10Styles.inkLt }}>
                {a.body}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* 하단 작은 푸터 */}
      <div style={{
        marginTop: 16, paddingTop: 10,
        borderTop: `1px solid ${v10Styles.ink}`,
        display: "flex", justifyContent: "space-between",
        fontFamily: v10Styles.sans, fontSize: 10,
        letterSpacing: "0.2em", color: v10Styles.inkLt,
      }}>
        <span>홈페이지 {data.website || ""}</span>
        <span>발행 {data.churchName}</span>
      </div>
    </div>
  );
}

function V10Inside({ data }) {
  const orderItems = data.order.filter(o => !o.kind);
  return (
    <div style={v10Styles.paper}>
      {/* 좌측 — 큰 인용문 페이지 (사진집의 텍스트 페이지) */}
      <div style={{
        flex: 0.85, padding: "60px 50px 40px",
        display: "flex", flexDirection: "column",
        justifyContent: "space-between",
        background: v10Styles.bgWhite,
        borderRight: `1px solid ${v10Styles.rule}`,
      }}>
        {/* 페이지 라벨 */}
        <div style={{
          fontFamily: v10Styles.sans, fontSize: 10,
          letterSpacing: "0.3em", color: v10Styles.inkLt,
        }}>
          ─ 본문 · {data.scripture}
        </div>

        {/* 큰 인용문 — 청자색 박스 */}
        <div style={{
          background: v10Styles.blueLt,
          padding: "32px 30px",
          borderLeft: `4px solid ${v10Styles.blue}`,
        }}>
          <div style={{
            fontFamily: v10Styles.serif, fontWeight: 400,
            fontSize: 36, lineHeight: 1.55, letterSpacing: "0.02em",
            color: v10Styles.blue,
          }}>
            {data.verseQuote.text}
          </div>
          <div style={{
            marginTop: 20,
            fontFamily: v10Styles.sans, fontSize: 11.5,
            letterSpacing: "0.25em", color: v10Styles.coral,
            fontWeight: 700,
          }}>
            ─ {data.verseQuote.ref}
          </div>
        </div>

        {/* 설교 정보 — 매우 작게 */}
        <div style={{
          marginTop: 28, paddingTop: 14,
          borderTop: `1px solid ${v10Styles.rule}`,
          display: "grid", gridTemplateColumns: "auto 1fr",
          columnGap: 18, rowGap: 3,
          fontFamily: v10Styles.sans, fontSize: 11.5,
        }}>
          <span style={{ color: v10Styles.inkLt, letterSpacing: "0.15em" }}>설교</span>
          <span style={{ fontWeight: 600 }}>{data.sermonTitle}</span>
          <span style={{ color: v10Styles.inkLt, letterSpacing: "0.15em" }}>설교자</span>
          <span>{data.pastor}</span>
          <span style={{ color: v10Styles.inkLt, letterSpacing: "0.15em" }}>일자</span>
          <span>{data.date}</span>
        </div>
      </div>

      {/* 우측 — 예배 순서 (단순한 세로 리스트, 라운드 모서리 / 카드 없음) */}
      <div style={{
        flex: 1.15, padding: "60px 50px 40px",
        display: "flex", flexDirection: "column",
        background: v10Styles.bg,
      }}>
        <div style={{
          display: "flex", alignItems: "baseline", justifyContent: "space-between",
          marginBottom: 20, paddingBottom: 8,
          borderBottom: `2px solid ${v10Styles.blue}`,
        }}>
          <div style={{ fontFamily: v10Styles.serif, fontSize: 28, fontWeight: 700, color: v10Styles.blue }}>
            오전 예배 순서
          </div>
          <div style={{
            fontFamily: v10Styles.sans, fontSize: 11,
            letterSpacing: "0.25em", color: v10Styles.inkLt,
          }}>
            {orderItems.length}개 순서
          </div>
        </div>

        <div style={{ flex: 1, fontFamily: v10Styles.sans, fontSize: 12.5 }}>
          {orderItems.map((o, i) => (
            <div key={i} style={{
              display: "grid",
              gridTemplateColumns: "26px 88px 1fr 90px",
              columnGap: 12,
              padding: "6px 0",
              alignItems: "baseline",
              borderBottom: `1px solid ${v10Styles.rule}`,
            }}>
              <span style={{
                fontVariantNumeric: "tabular-nums",
                color: v10Styles.inkLt, fontSize: 11,
              }}>
                {String(i + 1).padStart(2, "0")}
              </span>
              <span style={{ fontFamily: v10Styles.serif, fontWeight: 700, fontSize: 13.5 }}>
                {o.item}
              </span>
              <span style={{
                color: v10Styles.ink,
                fontSize: o.emphasize ? 13 : 12,
                fontFamily: o.emphasize ? v10Styles.serif : v10Styles.sans,
                fontWeight: o.emphasize ? 600 : 400,
              }}>
                {o.content || ""}
              </span>
              <span style={{
                fontSize: 10.5, color: v10Styles.inkLt,
                textAlign: "right", letterSpacing: "0.05em",
              }}>
                {o.who}
              </span>
            </div>
          ))}
        </div>

        {/* 봉사자 + 그 외 예배 — 단정한 푸터 */}
        <div style={{
          marginTop: 18, paddingTop: 12,
          borderTop: `1px solid ${v10Styles.ink}`,
          fontFamily: v10Styles.sans,
        }}>
          <div style={{ fontSize: 10.5, letterSpacing: "0.2em", color: v10Styles.inkLt, marginBottom: 6 }}>
            봉사자
          </div>
          <div style={{ fontSize: 12, lineHeight: 1.65 }}>
            {data.helpers.map((h, i) => (
              <span key={i}>
                <span style={{ color: v10Styles.inkLt }}>{h.role}</span>{" "}
                <span style={{ fontWeight: 600 }}>{h.name}</span>
                {i < data.helpers.length - 1 ? " 　 " : ""}
              </span>
            ))}
          </div>

          <div style={{
            marginTop: 16,
            display: "grid", gridTemplateColumns: "1fr 1fr",
            columnGap: 24,
          }}>
            {[
              { title: "오후 예배", items: data.afternoon },
              { title: "수요 예배", items: data.wednesday },
            ].map((sec, idx) => (
              <div key={idx}>
                <div style={{
                  fontFamily: v10Styles.serif, fontSize: 16, fontWeight: 700,
                  marginBottom: 5,
                }}>
                  {sec.title}
                </div>
                <div style={{ fontSize: 11.5, lineHeight: 1.7 }}>
                  {sec.items.map((o, j) => (
                    <div key={j} style={{
                      display: "flex", justifyContent: "space-between",
                      borderBottom: `0.5px solid ${v10Styles.rule}`,
                    }}>
                      <span>{o.item}</span>
                      <span style={{ color: v10Styles.inkLt }}>{o.content || "—"}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function V10Outside({ data }) {
  return (
    <div style={v10Styles.paper}>
      <V10Back data={data} />
      <V10Cover data={data} />
    </div>
  );
}

window.V10Outside = V10Outside;
window.V10Inside = V10Inside;
