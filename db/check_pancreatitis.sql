-- 1. Сколько признаков в каждом статусе. Ожидается: опубликован 11, черновик 1, удален 1
SELECT status, COUNT(*) FROM pancreatitis_signs GROUP BY status ORDER BY status;

-- 2. Всего лайков. Ожидается: 36
SELECT COUNT(*) FROM pancreatitis_sign_likes;

-- 3. Лайки по признакам (то, что потом покажет приложение). Ожидается 3 или 4 у каждого опубликованного
SELECT s.id, s.title, COUNT(l.id) AS likes
FROM pancreatitis_signs s
LEFT JOIN pancreatitis_sign_likes l ON l.sign_id = s.id
GROUP BY s.id, s.title
ORDER BY s.id;

-- 4. Ограничения таблицы признаков: внешний ключ на создателя (без ON DELETE CASCADE) и проверка статуса
SELECT conname, pg_get_constraintdef(oid) AS definition
FROM pg_constraint
WHERE conrelid = 'pancreatitis_signs'::regclass;

-- 5. Ограничения таблицы лайков: два внешних ключа и уникальная пара
SELECT conname, pg_get_constraintdef(oid) AS definition
FROM pg_constraint
WHERE conrelid = 'pancreatitis_sign_likes'::regclass;

-- 6. Индексы таблицы признаков: среди них уникальный индекс "не более одного черновика на пользователя"
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'pancreatitis_signs';
