import { Dispatch, SetStateAction, useState } from "react";
import { styled } from "styled-components";
import initialData from "../data.json"; // JSON 데이터 가져오기

interface WorshipItem {
  title: string;
  obj: string;
  info: string;
  lead?: string;
  children?: WorshipItem[];
}

interface Drag {
  target?: WorshipItem;
  to?: WorshipItem;
  grab: boolean;
  list: WorshipItem[];
}

const MoveEventWrap = styled.div`
  gap: 40px;
  margin-top: 40px;
`;

const StaticList = styled.div`
  width: 500px;
  display: flex;
  flex-wrap: wrap;
  flex-direction: row;
  gap: 5px;
  .static_el {
    padding: 8px;
    border-radius: 5px;
    border: 2px solid #28a745;
    color: #28a745;
    font-weight: bold;
    cursor: pointer;
    text-align: center;
    &:hover {
      background-color: #28a745;
      color: white;
    }
  }
`;

const DraggableList = styled.div`
  width: 300px;
  display: flex;
  flex-direction: column;
  max-height: 1000px;
  gap: 5px;
  .moveEvent_el {
    padding: 8px;
    border-radius: 5px;
    border: 2px solid #005cad;
    color: #005cad;
    font-weight: bold;
    cursor: grab;
    text-align: center;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-direction: column;
    &:active {
      cursor: grabbing;
      background-color: #005cad;
      color: #fff;
    }
    .delete_btn {
      background: red;
      color: white;
      border: none;
      padding: 4px 8px;
      cursor: pointer;
      font-size: 12px;
      margin-top: 5px;
      border-radius: 4px;
      &:hover {
        background: darkred;
      }
    }
    .expand_content {
      width: 100%;
      margin-top: 5px;
      padding: 8px;
      border-top: 1px solid #ccc;
      background-color: #f8f9fa;
      font-size: 14px;
      color: #333;
      text-align: left;
    }
  }
`;

const WorshipElement = ({
  dragging: [drag, setDrag],
  el,
  handleRemove,
  i,
  toggleExpand,
  expanded,
}: {
  dragging: [Drag, Dispatch<SetStateAction<Drag>>];
  el: WorshipItem;
  handleRemove: (item: WorshipItem) => void;
  i: number;
  toggleExpand: (title: string) => void;
  expanded: boolean;
}) => {
  return (
    <div
      className="moveEvent_el"
      key={el.title}
      draggable
      onDragStart={() => {
        setDrag({ ...drag, target: el, grab: true });
      }}
      onDragOver={(e) => {
        e.preventDefault();
        setDrag({ ...drag, to: el });
      }}
      onDrop={(e) => {
        e.preventDefault();
        const copyArr = [...drag.list];
        const [tg_ind, t_ind] = [
          copyArr.indexOf(drag.target!),
          copyArr.indexOf(drag.to!),
        ];
        const [slice] = copyArr.splice(tg_ind, 1);
        copyArr.splice(t_ind, 0, slice);
        setDrag({ ...drag, list: copyArr, grab: false });
      }}
      onDragEnd={() => {
        document.body.style.cursor = "default";
      }}
      onClick={() => toggleExpand(el.title)} // 클릭 시 펼치기/접기
    >
      {i + 1}. {el.title.split("_")[1]}
      {expanded && el.info.includes("edit") && (
        <div className="expand_content">
          <p>obj: {el.obj}</p>
          {el.lead && <p>lead: {el.lead}</p>}
        </div>
      )}
      <button className="delete_btn" onClick={(e) => { e.stopPropagation(); handleRemove(el); }}>
        삭제
      </button>
    </div>
  );
};

export default function MoveEvent() {
  const [staticList, setStaticList] = useState<WorshipItem[]>(initialData);
  const [drag, setDrag] = useState<Drag>({ list: [], grab: false });
  const [expandedItems, setExpandedItems] = useState<string[]>([]);

  const handleSelect = (item: WorshipItem) => {
    setStaticList(staticList.filter((el) => el.title !== item.title));
    setDrag({ ...drag, list: [...drag.list, item] });
  };

  const handleRemove = (item: WorshipItem) => {
    setDrag({
      ...drag,
      list: drag.list.filter((el) => el.title !== item.title),
    });
    setStaticList([...staticList, item]);
  };

  const toggleExpand = (title: string) => {
    setExpandedItems((prev) =>
      prev.includes(title) ? prev.filter((t) => t !== title) : [...prev, title]
    );
  };

  return (
    <MoveEventWrap>
      <h3>선택 목록</h3>
      <StaticList>
        {staticList.map((el) => (
          <div key={el.title} className="static_el" onClick={() => handleSelect(el)}>
            {el.title.split("_")[1]}
          </div>
        ))}
      </StaticList>

      <h3>예배 순서</h3>
      <DraggableList>
        {drag.list.map((el, i) => {
          const props = {
            dragging: [drag, setDrag] as [Drag, Dispatch<SetStateAction<Drag>>],
            el,
            handleRemove,
            i,
            toggleExpand,
            expanded: expandedItems.includes(el.title),
          };
          return <WorshipElement key={el.title} {...props} />;
        })}
      </DraggableList>
    </MoveEventWrap>
  );
}
