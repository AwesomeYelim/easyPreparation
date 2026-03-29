-- 다중 성경 버전 등록
INSERT INTO bible_versions (id, name, code) VALUES
  (2, '개역한글', 'KRV'),
  (3, '공동번역', 'COMMON'),
  (4, '표준새번역', 'NKSV'),
  (5, 'NIV', 'NIV'),
  (6, 'KJV', 'KJV'),
  (7, '우리말성경', 'URNB')
ON CONFLICT (id) DO NOTHING;
