class SeasonDecorator < ApplicationDecorator
  def yearly_season_ja
    seasons = {
      winter: "冬季",
      spring: "春季",
      summer: "夏季",
      autumn: "秋季"
    }

    "#{year}年#{seasons[name.to_sym]}"
  end
end
