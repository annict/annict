# frozen_string_literal: true

module SeasonsHelper
  def season_local_name(year, name)
    t("resources.season.yearly.#{name}", year: year)
  end
end
