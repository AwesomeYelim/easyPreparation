import { useSetRecoilState, useRecoilState } from "recoil";
import { worshipOrderState, selectedDetailState } from "@/recoilState";
import React from "react";

export default function SelectedOrder({ selectedItems, setSelectedItems }) {
  // setWorshipOrder를 useSetRecoilState로 변경하여 setter만 사용합니다.
  const setWorshipOrder = useSetRecoilState(worshipOrderState);
  const [selectedDetail, setSelectedDetail] =
    useRecoilState(selectedDetailState);

  const handleDeleteItem = (item) => {
    setSelectedItems((prevItems) => prevItems.filter((el) => el !== item));
    setWorshipOrder((prevOrder) => [...prevOrder, item]);
  };

  return (
    <section className="card">
      <h2>선택된 예배 순서</h2>
      <div>
        {selectedItems.map((item) => {
          // 각 항목에 대한 상세 정보를 selectedDetail 객체에서 key로 접근
          return (
            <>
              <span className="tag" onClick={() => setSelectedDetail(item)}>
                {item.title.split("_")[1]}
                <button
                  className="delete-btn"
                  onClick={() => handleDeleteItem(item)}
                >
                  ❌
                </button>
              </span>
              {/* <div key={item} className="detail-card">
                <p>
                  <strong>Obj:</strong> {item.obj}
                </p>
                <p>
                  <strong>Info:</strong> {item.info}
                </p>
                <p>
                  <strong>Lead:</strong> {item.lead}
                </p>
              </div> */}
            </>
          );
        })}
      </div>
      <>
        {selectedDetail && (
          <div key={selectedDetail.title} className="detail-card">
            <p>
              <strong>Obj:</strong> {selectedDetail.obj}
            </p>
            <p>
              <strong>Info:</strong> {selectedDetail.info}
            </p>
            <p>
              <strong>Lead:</strong> {selectedDetail.lead}
            </p>
          </div>
        )}
      </>
    </section>
  );
}
