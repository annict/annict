FactoryGirl.define do
  factory :draft_work do
    sequence(:title) { |n| "#{n}人はプリキュア" }
    sequence(:title_kana) { |n| "#{n}にんはぷりきゅあ" }
    media :tv
    official_site_url "http://example.com"
    wikipedia_url "http://example.com"
  end
end
