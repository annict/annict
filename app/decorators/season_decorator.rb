# frozen_string_literal: true

class SeasonDecorator < ApplicationDecorator
  def yearly_season_ja
    return "#{year}年" if name == "all"
    "#{year}年#{Season::NAME_DATA[name.to_sym]}"
  end

  def humanize_name
    I18n.t("resources.season.yearly.#{name}", year: year)
  end
end
