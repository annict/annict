FactoryGirl.define do
  factory :season do
    sequence(:year) { |n| "20#{sprintf('%02d', n)}" }
    name "winter"
    sequence(:sort_number) { |n| n }
    sequence(:slug) { |n| "20#{sprintf('%02d', n)}-spring" }
  end
end
