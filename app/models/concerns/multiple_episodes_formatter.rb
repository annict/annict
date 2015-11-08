require "csv"

module MultipleEpisodesFormatter
  extend ActiveSupport::Concern

  included do
    def to_episode_hash
      # ダブルクォートを含んだ文字列を `CSV.parse` すると `CSV::MalformedCSVError` が
      # 発生するので、パース前に置換する
      body = self.body.gsub(/"/, "__double_quote__")

      CSV.parse(body).map do |ary|
        title = ary[1].gsub("__double_quote__", '"').try!(:strip) if ary[1].present?
        { number: ary[0].try!(:strip), title: title }
      end
    end
  end
end
