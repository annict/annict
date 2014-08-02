FactoryGirl.define do
  factory :user do
    sequence(:username) { |n| "user_#{n}" }
    sequence(:email)    { |n| "user_#{n}@example.com" }

    trait :with_profile do
      after :create do |user|
        create(:profile, { user: user })
      end
    end

    trait :with_provider do
      after :create do |user|
        create(:provider, { user: user })
      end
    end

    factory :registered_user, traits: [:with_profile, :with_provider] do
      after :create, &:confirm!
    end
  end
end
