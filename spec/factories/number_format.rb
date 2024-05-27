# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :number_format do
    name { "第1話" }
    format { "第%d話" }
  end
end
