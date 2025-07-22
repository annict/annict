# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :user do
    sequence(:username) { |n| "user_#{n}" }
    sequence(:email) { |n| "user_#{n}@example.com" }
    time_zone { "Asia/Tokyo" }
    locale { "ja" }

    trait :with_profile do
      after(:create) do |user|
        create(:profile, user: user)
      end
    end

    trait :with_provider do
      after :create do |user|
        create(:provider, user: user)
      end
    end

    trait :with_setting do
      after :create do |user|
        create(:setting, user: user)
      end
    end

    trait :with_email_notification do
      after :create do |user|
        create(:email_notification, user: user)
      end
    end

    # 編集者権限
    trait :with_editor_role do
      role { :editor }
    end

    # 管理者権限
    trait :with_admin_role do
      role { :admin }
    end

    # Annictサポーター
    trait :with_supporter do
      gumroad_subscriber
    end

    factory :registered_user, traits: %i[with_profile with_provider with_setting with_email_notification] do
      after :create do |user|
        user.confirm
      end
    end
  end
end
