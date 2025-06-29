import { WorshipOrderItem } from "../page";

export function ResultPart({
  selectedItems,
}: {
  selectedItems: WorshipOrderItem[];
}) {
  function formatBibleReference(obj: string): string {
    const bibleRegex = /^(.+?)_\d+\/(\d+):(\d+)(?:-(\d+):)?(\d+)?$/;

    return obj
      .split(",")
      .map((item) => {
        const trimmed = item.trim();
        const match = trimmed.match(bibleRegex);
        if (!match) return trimmed;

        const [_, bookName, chapterStart, verseStart, chapterEnd, verseEnd] =
          match;

        // 단일절
        if (!chapterEnd && !verseEnd) {
          return `${bookName} ${chapterStart}:${verseStart}`;
        }

        // 절 범위, 같은 장
        if (!chapterEnd || chapterStart === chapterEnd) {
          return `${bookName} ${chapterStart}:${verseStart}-${verseEnd}`;
        }

        // 절 범위, 다른 장
        return `${bookName} ${chapterStart}:${verseStart}-${chapterEnd}:${verseEnd}`;
      })
      .join(", ");
  }
  console.log(selectedItems);
  return (
    <div className="card">
      <h2>생성된 예배 내용</h2>
      <div className="contents">
        {(() => {
          const result = [];

          for (const el of selectedItems) {
            if (el.title !== "말씀내용" && el.title !== "행사") {
              result.push(
                <div className="row" key={el.title + el.obj}>
                  <div className="title">{el.title}</div>
                  <div
                    className="obj"
                    title={
                      el.info === "b_edit"
                        ? formatBibleReference(el.obj)
                        : el.obj
                    }
                  >
                    {el.info === "b_edit"
                      ? formatBibleReference(el.obj)
                      : el.obj}
                  </div>
                  <div className="lead">{el.lead}</div>
                </div>
              );
            }

            if (el.title === "축도") break;
          }

          return result;
        })()}
      </div>
    </div>
  );
}
