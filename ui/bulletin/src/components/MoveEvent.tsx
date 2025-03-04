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
  display: flex;
  flex-direction: row;
  gap: 40px;
  margin-top: 40px;
`;

const StaticList = styled.div`
  width: 300px;
  max-height: 1000px;
  display: flex;
  flex-direction: column;
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
    &:active {
      cursor: grabbing;
    }
    .delete_btn {
      background: red;
      color: white;
      border: none;
      padding: 4px 8px;
      cursor: pointer;
      font-size: 12px;
      margin-left: 10px;
      border-radius: 4px;
      &:hover {
        background: darkred;
      }
    }
  }
`;

const WorshipElement = ({
  dragging: [drag, setDrag],
  el,
  handleRemove,
}: {
  dragging: [Drag, Dispatch<SetStateAction<Drag>>];
  el: WorshipItem;
  handleRemove: (item: WorshipItem) => void;
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
        const [tg_ind, t_ind] = [copyArr.indexOf(drag.target!), copyArr.indexOf(drag.to!)];
        const [slice] = copyArr.splice(tg_ind, 1);
        copyArr.splice(t_ind, 0, slice);
        setDrag({ ...drag, list: copyArr, grab: false });
      }}
      onDragEnd={() => {
        document.body.style.cursor = "default";
      }}>
      {el.title} - {el.obj}
      <button className="delete_btn" onClick={() => handleRemove(el)}>
        삭제
      </button>
    </div>
  );
};

export default function MoveEvent() {
  const [staticList, setStaticList] = useState<WorshipItem[]>(initialData);
  const [drag, setDrag] = useState<Drag>({ list: [], grab: false });

  const handleSelect = (item: WorshipItem) => {
    setStaticList(staticList.filter((el) => el.title !== item.title));
    setDrag({ ...drag, list: [...drag.list, item] });
  };

  const handleRemove = (item: WorshipItem) => {
    setDrag({ ...drag, list: drag.list.filter((el) => el.title !== item.title) });
    setStaticList([...staticList, item]); // 원래 목록으로 복귀
  };

  return (
    <MoveEventWrap>
      {/* 선택 가능한 정적 목록 */}
      <StaticList>
        <h3>선택 목록</h3>
        {staticList.map((el) => (
          <div key={el.title} className="static_el" onClick={() => handleSelect(el)}>
            {el.title} - {el.obj}
          </div>
        ))}
      </StaticList>

      {/* 드래그 & 드롭 가능한 목록 */}
      <DraggableList>
        <h3>예배 순서</h3>
        {drag.list.map((el) => {
          const props = {
            dragging: [drag, setDrag] as [Drag, Dispatch<SetStateAction<Drag>>],
            el,
            handleRemove,
          };
          return <WorshipElement key={el.title} {...props} />;
        })}
      </DraggableList>
    </MoveEventWrap>
  );
}
