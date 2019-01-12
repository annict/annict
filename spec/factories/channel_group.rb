# frozen_string_literal: true

FactoryBot.define do
  factory :channel_group do
    sequence(:sc_chgid)
    name { "テレビ 関東" }
  end
end
