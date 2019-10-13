# frozen_string_literal: true

FactoryBot.define do
  factory :trailer do
    work
    url { "https://www.youtube.com/watch?v=2ZR6fCnPcvA" }
    title { "コミックマーケット86公開PV" }
  end
end
