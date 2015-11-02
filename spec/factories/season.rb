FactoryGirl.define do
  factory :season do
    sequence(:year) { |n| format("20%02d", n).to_i }
    name "winter"
    sequence(:sort_number) { |n| n }
  end
end
