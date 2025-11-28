# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :work_record do
    association :user, :with_profile
    work
    record
    body { "おもしろかった" }
    locale { "ja" }
  end
end
