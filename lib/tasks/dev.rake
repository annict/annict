# typed: false
# frozen_string_literal: true

namespace :dev do
  # データベースをリストアする
  # 使用例:
  # docker compose exec -e ANNICT_POSTGRES_DUMP_PATH=tmp/db.dump app bin/rails dev:restore_db
  task restore_db: :environment do
    sql = "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
    ActiveRecord::Base.connection.execute(sql)

    system "
      pg_restore \
        --verbose \
        --clean \
        --no-acl \
        --no-owner \
        --jobs 4 \
        -h #{ENV.fetch("ANNICT_POSTGRES_HOST")} \
        -U #{ENV.fetch("ANNICT_POSTGRES_USERNAME")} \
        -d annict_development \
        -p #{ENV.fetch("ANNICT_POSTGRES_PORT")} \
        #{ENV.fetch("ANNICT_POSTGRES_DUMP_PATH")}
    "
  end
end
