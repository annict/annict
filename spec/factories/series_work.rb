# frozen_string_literal: true

FactoryBot.define do
  factory :series_work do
    series
    work
    summary { "TVシリーズ" }
  end
end
