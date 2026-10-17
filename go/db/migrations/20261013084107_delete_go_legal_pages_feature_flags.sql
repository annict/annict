-- migrate:up

-- 法的文書のページをフラグに関係なくGo版で配信するようにしたため、go_legal_pagesフラグの行は参照されなくなった。
DELETE FROM feature_flags WHERE name = 'go_legal_pages';

-- migrate:down

-- 削除した行はどの利用者・端末に付いていたかを復元できないため、何もしない。
