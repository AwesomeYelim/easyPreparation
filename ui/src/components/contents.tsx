import React, { useState } from "react";

interface Info {
  title: string;
  obj: string;
  info: string;
  children?: Info[]; // children 속성을 추가
}

const worshipTitle = "main_worship";
const initialData: Info[] = [
  {
    title: "1_전주",
    obj: "전주",
    info: "l_전주",
  },
  {
    title: "2_예배의 부름",
    obj: "렘 33:2-3",
    info: "c_edit",
  },
  {
    title: "3_찬송",
    obj: "8장",
    info: "c_edit",
  },
  {
    title: "4_성시교독",
    obj: "3. 시편 4편",
    info: "c_edit",
  },
  {
    title: "5_신앙고백",
    obj: "사도신경",
    info: "c_사도신경",
  },
  {
    title: "6_찬송",
    obj: "259장",
    info: "c_edit",
  },
  {
    title: "7_기도",
    obj: "이경아 사모",
    info: "r_edit",
  },
  {
    title: "8_성경봉독",
    obj: "고후 1:11",
    info: "c_edit",
  },
  {
    title: "8.1_말씀내용",
    obj: "말씀내용",
    info: "c_edit",
  },
  {
    title: "9_찬양",
    obj: "찬양",
    info: "l_찬양",
  },
  {
    title: "10_참회의 기도",
    obj: "참회의 기도",
    info: "l_참회의 기도",
  },
  {
    title: "11_말씀",
    obj: "감사의 감사를 낳는 기도",
    info: "c_edit",
  },
  {
    title: "12_헌금봉헌",
    obj: "365장",
    info: "c_edit",
  },
  {
    title: "13_교회소식",
    obj: "교회소식",
    info: "edit",
    children: [
      {
        title: "예배 참여 안내",
        obj: "매주 금요일 나라와 민족을 위하여 기도 하고 있습니다. \n많은 참여바랍니다. (아래 예배시간 참고)",
        info: "c_edit",
      },
      {
        title: "교회절기 및 행사",
        obj: "",
        info: "",
        children: [
          {
            title: "새벽기도",
            obj: "특별새벽기도회를 은혜가운데 잘 마쳤습니다. (기도로 2025년도 승리합시다.)",
            info: "c_edit",
          },
          {
            title: "예배 후",
            obj: "예. 결산 공동의회: 1/19(주일) 예배 후에 - 다음주",
            info: "c_edit",
          },
        ],
      },
      {
        title: "담임 목사 활동",
        obj: "목사합창단 모임 : 1/21(화) 오전 11시, 동산교회당",
        info: "c_edit",
      },
      {
        title: "선교회 소식",
        obj: "연말정산용 기부금 납부 증명서 신청받습니다. - 재정부장님께",
        info: "c_edit",
      },
      {
        title: "노회 소식",
        obj: "노회임원 정치부 연석모임 : 1/23(목) 오전 11시, 노회 사무실",
        info: "c_edit",
      },
    ],
  },
  {
    title: "14_찬송",
    obj: "635장",
    info: "c_edit",
  },
  {
    title: "15_내주기도",
    obj: "이병용 집사",
    info: "edit",
  },
  {
    title: "16_헌금, 안내",
    obj: "남선교회",
    info: "edit",
  },
  {
    title: "17_오늘의 말씀",
    obj: "창 4:3-4",
    info: "edit",
  },
];

const EditableData: React.FC = () => {
  const [title, setTitle] = useState(worshipTitle);
  const [data, setData] = useState(initialData);

  const handleInputChange = (key: string, newObj: string) => {
    const updateData = (items: Info[], keyParts: string[]): Info[] => {
      const [currentIndex, ...restKeyParts] = keyParts;

      if (!currentIndex) return items;

      return items.map((item, index) => {
        if (index === parseInt(currentIndex)) {
          if (restKeyParts.length === 0) {
            // 최종 obj 업데이트
            return { ...item, obj: newObj };
          }
          if (item.children) {
            // children 업데이트
            return {
              ...item,
              children: updateData(item.children, restKeyParts),
            };
          }
        }
        return item;
      });
    };

    const keyParts = key.split("-");
    setData((prevData) => updateData(prevData, keyParts));
  };

  const handleSubmit = async () => {
    try {
      // 실제 Go와 연동 시 사용될 함수 호출
      if (window.sendContentsDate) {
        await window.sendContentsDate(title, data);
      }
    } catch (error) {
      console.error("Error:", error);
    }
  };

  // 재귀적으로 요소를 렌더링하는 함수
  const renderItems = (items: Info[], parentIndex: string = "") => {
    return items.map((item, index) => {
      const key = parentIndex ? `${parentIndex}-${index}` : `${index}`;

      return (
        <div key={key} style={{ marginBottom: "15px" }}>
          <label
            style={{
              marginTop: "10px",
              color: item.info.includes("edit") ? "#000000" : "#ccc",
            }}
          >
            {item.title}
          </label>
          {item.info.includes("edit") && (
            <input
              type="text"
              onChange={(e) => handleInputChange(key, e.target.value)}
              placeholder={item.obj}
              style={{
                marginTop: "5px",
                display: "block",
                padding: "10px",
                width: "100%",
                maxWidth: "400px",
                border: "1px solid #ccc",
                borderRadius: "4px",
              }}
            />
          )}
          {/* children이 있으면 재귀 호출 */}
          {item.children && (
            <div style={{ marginLeft: "20px", marginTop: "10px" }}>
              {renderItems(item.children, key)}
            </div>
          )}
        </div>
      );
    });
  };
  return (
    <div>
      <div style={{ marginBottom: "15px" }}>
        <label>예배 종류</label>
        <input
          type="text"
          onChange={(e) => setTitle(e.target.value)}
          placeholder={worshipTitle}
          style={{
            display: "block",
            marginTop: "10px",
            padding: "10px",
            width: "100%",
            maxWidth: "400px",
            border: "1px solid #ccc",
            borderRadius: "4px",
          }}
        />
      </div>

      {/* 재귀적으로 데이터를 렌더링 */}
      {renderItems(data)}

      <button
        onClick={handleSubmit}
        style={{
          padding: "10px 20px",
          backgroundColor: "#4CAF50",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: "pointer",
        }}
      >
        Submit
      </button>
    </div>
  );
};

export default EditableData;
