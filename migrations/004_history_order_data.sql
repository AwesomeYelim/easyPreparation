-- 004: generation_history에 order_data TEXT 컬럼 추가
-- Display 전송 이력에 예배 순서 데이터를 저장하여 재사용 가능하게 함

ALTER TABLE generation_history ADD COLUMN order_data TEXT;
