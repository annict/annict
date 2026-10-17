-- migrate:up

-- 退会の画面をフラグに関係なくGo版で配信するようにしたため、go_account_deletionフラグの行は参照されなくなった。
DELETE FROM feature_flags WHERE name = 'go_account_deletion';

-- migrate:down

-- 削除した行はどの利用者・端末に付いていたかを復元できないため、何もしない。
