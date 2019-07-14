# frozen_string_literal: true

module Localable
  private

  def locale_ja?
    locale.to_s == "ja"
  end

  def locale_en?
    locale.to_s == "en"
  end
end
