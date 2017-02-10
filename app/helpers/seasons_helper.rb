# frozen_string_literal: true

module SeasonsHelper
  def season_local_name(year, name, locale = nil)
    I18n.locale = locale if locale.present?
    t("resources.season.yearly.#{name}", year: year)
  end
end
