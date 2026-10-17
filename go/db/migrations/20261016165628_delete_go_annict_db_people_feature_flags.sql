-- migrate:up

-- 人物のAnnict DBをフラグに関係なくGo版で配信するようにしたため、go_annict_db_peopleフラグの行は参照されなくなった。
DELETE FROM feature_flags WHERE name = 'go_annict_db_people';

-- migrate:down

-- 削除した行はどの利用者・端末に付いていたかを復元できないため、何もしない。
