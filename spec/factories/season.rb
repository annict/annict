FactoryGirl.define do
  factory :season do
    sequence(:year) { |n| "20%02d-spring" % n }
    name "winter"
    sequence(:sort_number) { |n| n }
    sequence(:slug) { |n| "20%02d-spring" % n }
  end
end
