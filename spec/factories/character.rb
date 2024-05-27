# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :character do
    series
    sequence(:name) { |n| "山田#{n}郎" }
    sequence(:name_kana) { |n| "やまだ#{n}ろう" }
    sequence(:name_en) { |n| "Yamada, #{n}rou" }
    nickname { "nickname_data" }
    nickname_en { "nickname_en_data" }
    birthday { "birthday_data" }
    birthday_en { "birthday_en_data" }
    age { "age_data" }
    age_en { "age_en_data" }
    blood_type { "blood_type_data" }
    blood_type_en { "blood_type_en_data" }
    height { "height_data" }
    height_en { "height_en_data" }
    weight { "weight_data" }
    weight_en { "weight_en_data" }
    nationality { "nationality_data" }
    nationality_en { "nationality_en_data" }
    occupation { "occupation_data" }
    occupation_en { "occupation_en_data" }
    description { "description_data" }
    description_en { "description_en_data" }
    description_source { "description_source_data" }
    description_source_en { "description_source_en_data" }

    trait :published do
      unpublished_at { nil }
    end

    trait :unpublished do
      unpublished_at { Time.zone.now }
    end

    trait :not_deleted do
      deleted_at { nil }
    end

    trait :deleted do
      deleted_at { Time.zone.now }
    end
  end
end
