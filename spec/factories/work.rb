FactoryGirl.define do
  factory :work do
    sequence(:title) { |n| "#{n}人はプリキュア" }
    media :tv
    official_site_url 'http://example.com'
    wikipedia_url 'http://example.com'
    on_air false

    trait :on_air do
      on_air true
    end

    trait :with_item do
      after :create do |work|
        create(:item, { work: work })
      end
    end
  end
end
