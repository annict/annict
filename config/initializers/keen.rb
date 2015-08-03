# 開発環境でKeenIO関係の環境変数が定義されていないときは
# `Keen.publish` が機能しないようにする
if Rails.env.development? && ENV["KEEN_PROJECT_ID"].blank?
  Keen.class_eval do
    def self.publish(event_collection, properties); end
  end
end
