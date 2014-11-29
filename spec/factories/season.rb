FactoryGirl.define do
  factory :season do
    sequence(:name) { |n| "20#{sprintf('%02d', n)}年春季" }
    sequence(:slug) { |n| "20#{sprintf('%02d', n)}-spring" }
  end
end
