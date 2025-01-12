import React, { useState } from 'react';

const initialData = [
        {
            "title": "1_전주",
            "content": "전주",
            "info": "l_전주"
        },
        {
            "title": "2_예배의 부름",
            "content": "하 2:20",
            "info": "c_edit"
        },
        {
            "title": "3_찬송",
            "content": "5장",
            "info": "c_edit"
        },
        {
            "title": "4_성시교독",
            "content": "56. 시편 128편",
            "info": "c_edit"
        },
        {
            "title": "5_신앙고백",
            "content": "사도신경",
            "info": "c_사도신경"
        },
        {
            "title": "6_찬송",
            "content": "257장",
            "info": "c_edit"
        },
        {
            "title": "7_기도",
            "content": "홍영란권사",
            "info": "r_edit"
        },
        {
            "title": "8_성경봉독",
            "content": "시 84:1-12",
            "info": "c_edit"
        },
        {
            "title": "8.1_말씀내용",
            "content": "말씀내용",
            "info": "c_edit"
        },
        {
            "title": "9_찬양",
            "content": "찬양",
            "info": "l_찬양"
        },
        {
            "title": "10_참회의 기도",
            "content": "참회의 기도",
            "info": "l_참회의 기도"
        },
        {
            "title": "11_말씀",
            "content": "시온의 대로가 있는 자",
            "info": "c_edit"
        },
        {
            "title": "12_헌금봉헌",
            "content": "208장",
            "info": "c_edit"
        },
        {
            "title": "13_교회소식",
            "content": "교회소식",
            "info": "l_교회소식"
        },
        {
            "title": "14_찬송",
            "content": "635장",
            "info": "c_edit"
        },
        {
            "title": "15_내주기도",
            "content": "이경아 권사님",
            "info": "edit"
        },
        {
            "title": "16_헌금, 안내",
            "content": "남선교회",
            "info": "edit"
        },
        {
            "title": "17_오늘의 말씀",
            "content": "세월이 지난 후에 가인은 땅의 소산으로 제물을 삼아 여호와께 드렸고 아벨은 자기도 양의 첫 새끼와 그 기름으로 드렸더니 여호와께서 아벨과 그의 제물은 받으셨으나\n\r\n창 4:3-4",
            "info": "edit"
        },
        {
            "title": "18.1_예배 참여 안내",
            "content": "매주 금요일 나라와 민족을 위하여 기도 하고 있습니다. \n많은 참여바랍니다. (아래 예배시간 참고)",
            "info": "c_edit"
        },
        {
            "title": "18.2_교회절기 및 행사",
            "content": "2) 예. 결산 공동의회: 12/19(주일) 예배 후에",
            "info": "c_2_edit"
        },
        {
            "title": "18.3_담임 목사 활동",
            "content": "정치부 모임: 1/23(목) 오전 11시, 노회 사무실",
            "info": "c_edit"
        },
        {
            "title": "18.4_선교회 소식",
            "content": "연말정산용 기부금 납부 증명서 신청받습니다. - 재정부장님께",
            "info": "c_edit"
        },
        {
            "title": "18.5_노회 소식",
            "content": "신년 하례회 - 1/7(화) 오전 11시, 영광교회",
            "info": "c_edit"
        }
    ]
;

const FigmaTokenInput: React.FC = () => {
  const [token, setToken] = useState("");
  const [key, setKey] = useState("");
  const [data, setData] = useState(initialData);
  const [responseMessage, setResponseMessage] = useState("");

  const handleInputChange = (index: number, newContent: string) => {
    const updatedData = [...data];
    updatedData[index].content = newContent;
    setData(updatedData);
  };

  const handleSubmit = async () => {
    if (!token || !key) {
      alert("Both token and key must be provided.");
      return;
    }

    try {
      // 실제 Go와 연동 시 사용될 함수 호출
      // await window.sendTokenAndKey(token, key);
      setResponseMessage("Success! Data sent to Go.");
    } catch (error) {
      console.error('Error:', error);
    }
  };

  return (
      <div style={{ margin: "20px", fontFamily: "Arial, sans-serif" }}>
        <h1>Figma Token and Key</h1>
        <div style={{ marginBottom: "20px" }}>
          <input
              type="text"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="Enter your token"
              style={{
                margin: "10px 0",
                padding: "10px",
                width: "100%",
                maxWidth: "400px",
                border: "1px solid #ccc",
                borderRadius: "4px",
              }}
          />
          <input
              type="text"
              value={key}
              onChange={(e) => setKey(e.target.value)}
              placeholder="Enter your key"
              style={{
                margin: "10px 0",
                padding: "10px",
                width: "100%",
                maxWidth: "400px",
                border: "1px solid #ccc",
                borderRadius: "4px",
              }}
          />
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
          <p>{responseMessage}</p>
        </div>

        <h1>Editable Data</h1>
        <div>
          {data.map((item, index) => (
              <div key={index} style={{ marginBottom: "15px" }}>
                <label>{item.title}</label>
                <input
                    type="text"
                    value={item.content}
                    onChange={(e) => handleInputChange(index, e.target.value)}
                    placeholder={item.content}
                    style={{
                      display: "block",
                      marginTop: "5px",
                      padding: "10px",
                      width: "100%",
                      maxWidth: "400px",
                      border: "1px solid #ccc",
                      borderRadius: "4px",
                    }}
                />
              </div>
          ))}
        </div>
      </div>
  );
};

export default FigmaTokenInput;
