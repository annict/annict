# typed: false
# frozen_string_literal: true

# rack-attackのカウンターを `rate_limits` テーブルに持つストア。
# Go版の限定器 (`internal/ratelimit`) と同じテーブルを、固定窓のupsertで読み書きする。
# テーブルの定義と期限切れの窓の削除はGo版が持つ。
#
# rack-attackは「`{prefix}:{窓の通し番号}:{throttleの名前}:{識別子}`」の形のキーを渡すため、
# 窓の通し番号に窓の長さを掛けて `window_start` を求め、残りを `key` とする。
# 窓の長さはキーに含まれないため、このストアを使うthrottleはすべて同じ長さの窓にする。
#
# カウンターの読み書きに失敗したときは、制限を掛けずに通すためnilを返す。
class RateLimitStore
  def initialize(period:, prefix: "rack::attack")
    @period = period.to_i
    @key_pattern = /\A#{Regexp.escape(prefix)}:(\d+):(.+)\z/
  end

  def increment(name, amount = 1, _options = {})
    key, window_start = parse(name)

    execute(:select_value, <<~SQL.squish, key, window_start, amount)&.to_i
      INSERT INTO rate_limits (key, window_start, count)
      VALUES (?, ?, ?)
      ON CONFLICT (key, window_start)
      DO UPDATE SET count = rate_limits.count + EXCLUDED.count
      RETURNING count
    SQL
  end

  def read(name, _options = {})
    key, window_start = parse(name)

    execute(:select_value, <<~SQL.squish, key, window_start)&.to_i
      SELECT count FROM rate_limits WHERE key = ? AND window_start = ?
    SQL
  end

  def write(name, value, _options = {})
    key, window_start = parse(name)

    # 増分の失敗後にrack-attackが呼ぶため、既存の件数は上書きしない。
    execute(:exec_update, <<~SQL.squish, key, window_start, value.to_i)
      INSERT INTO rate_limits (key, window_start, count)
      VALUES (?, ?, ?)
      ON CONFLICT (key, window_start)
      DO NOTHING
    SQL
  end

  def delete(name, _options = {})
    key, window_start = parse(name)

    execute(:exec_delete, <<~SQL.squish, key, window_start)
      DELETE FROM rate_limits WHERE key = ? AND window_start = ?
    SQL
  end

  private def parse(name)
    match = @key_pattern.match(name)
    raise ArgumentError, "窓の通し番号を含まないキーは扱えません: #{name}" unless match

    [match[2], Time.at(match[1].to_i * @period).utc]
  end

  private def execute(method, sql, *binds)
    ActiveRecord::Base.with_connection do |connection|
      connection.public_send(method, ActiveRecord::Base.sanitize_sql_array([sql, *binds]), "RateLimitStore")
    end
  rescue ActiveRecord::ActiveRecordError => e
    Rails.logger.warn("Rate limit store error: #{e.class}: #{e.message}")
    nil
  end
end
