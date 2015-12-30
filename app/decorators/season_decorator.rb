class SeasonDecorator < ApplicationDecorator
  def yearly_season_ja
    return "#{year}年" if name == "all"
    "#{year}年#{Season::NAME_DATA[name.to_sym]}"
  end
end
