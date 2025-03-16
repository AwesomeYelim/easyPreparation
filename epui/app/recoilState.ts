import { atom } from "recoil";

// 예배 순서 상태
export const worshipOrderState = atom({
  key: "worshipOrderState",
  default: [
    {
      title: "1_전주",
      obj: "전주",
      info: "-",
      lead: "반주자",
    },
    {
      title: "2_예배의 부름",
      obj: "렘 33:2-3",
      info: "b_edit",
      lead: "인도자",
    },
    {
      title: "3_개회기도",
      obj: "-",
      info: "-",
      lead: "인도자",
    },
    {
      title: "4_찬송",
      obj: "29장",
      info: "c_edit",
      lead: "인도자",
    },
    {
      title: "5_성시교독",
      obj: "125. 사순절(2)",
      info: "c_edit",
      lead: "일어서서",
    },
    {
      title: "6_신앙고백",
      obj: "사도신경",
      info: "-",
      lead: "일어서서",
    },
    {
      title: "7_찬송",
      obj: "526장",
      info: "c_edit",
      lead: "함께",
    },
    {
      title: "8_대표기도",
      obj: "-",
      info: "r_edit",
      lead: "오남이 권사님",
    },
    {
      title: "9_성경봉독",
      obj: "고후 1:11",
      info: "b_edit",
      lead: "일어서서",
    },
    {
      title: "9.1_말씀내용",
      obj: "-",
      info: "c_edit",
    },
    {
      title: "10_찬양",
      obj: "-",
      info: "-",
      lead: "할렐루야성가대",
    },
    {
      title: "11_참회의 기도",
      obj: "먼저 우리 자신의 회개\n개인의 기도제목\n우리교회 신앙 공동체들을 위해\n교회의 재정 문제를 위해\n나라와 민족을 위해\n우리 주위의 아픈 환우들을 위해\n세계 복음화 사명",
      info: "edit",
      lead: "함께",
    },
    {
      title: "12_말씀",
      obj: "신령한 것을 먹고 마시는 사람들",
      info: "c_edit",
      lead: "홍은익목사",
    },
    {
      title: "13_기도",
      obj: "-",
      info: "-",
      lead: "인도자",
    },
    {
      title: "0_성찬예식",
      obj: "-",
      info: "-",
      lead: "함께",
    },
    {
      title: "14_헌금봉헌",
      obj: "143장",
      info: "c_edit",
      lead: "함께",
    },
    {
      title: "15_봉헌기도",
      obj: "-",
      info: "-",
      lead: "인도자",
    },
    {
      title: "16_교회소식",
      obj: "",
      info: "edit",
      lead: "인도자",
      children: [
        {
          title: "예배 참여 안내",
          obj: "매주 금요일 나라와 민족을 위하여 기도 하고 있습니다. \n많은 참여바랍니다. (아래 예배시간 참고)",
          info: "c_edit",
        },
        {
          title: "교회절기 및 행사",
          obj: "-",
          info: "c_edit",
          children: [
            {
              title: "사순절 첫째 주일",
              obj: "3/9(주일)",
              info: "c_edit",
            },
            {
              title: "성경 통독(1차)",
              obj: "2/24(월) ~ 5/4(주일)/ 현 민수기 통독중",
              info: "c_edit",
            },
            {
              title: "춘계 대심방",
              obj: "3월 중순 부터 ~",
              info: "c_edit",
            },
          ],
        },
        {
          title: "노회 소식",
          obj: "-",
          info: "c_edit",
          children: [
            {
              title: "정치부",
              obj: "3/14(금) 오전 11시, 노회 사무실",
              info: "c_edit",
            },
            {
              title: "노회",
              obj: "4/22(화) 오전 9시 30분 ~ 오후 9시, 중부명성교회",
              info: "c_edit",
            },
          ],
        },
        {
          title: "교회 소식",
          obj: "매월 2째, 4째주는 교회에서 식사가 제공됩니다. (반찬 봉사할 가정 모집중입니다. 많은 참여 부탁드려요!!)",
          info: "c_edit",
        },
        {
          title: "오후 예배",
          obj: "오후 2시 예배 많은 참여 부탁드립니다.",
          info: "c_edit",
        },
      ],
    },
    {
      title: "17_찬송",
      obj: "635장",
      info: "c_edit",
      lead: "일어서서",
    },
    {
      title: "18_축도",
      obj: "-",
      info: "-",
      lead: "홍은익목사",
    },
    {
      title: "19_내주기도",
      obj: "홍영란 권사",
      info: "edit",
    },
    {
      title: "20_헌금, 안내",
      obj: "2여전도회",
      info: "edit",
    },
    {
      title: "21_오늘의 말씀",
      obj: "창 4:3-4",
      info: "b_edit",
    },
  ],
});

export const selectedDetailState = atom({
  key: "selectedDetailState",
  default: {}, // 선택한 항목의 상세 정보를 저장
});

// 교회 소식 상태
export const churchNewsState = atom({
  key: "churchNewsState",
  default: [],
});
