# typed: false
# frozen_string_literal: true

FactoryBot.define do
  factory :stripe_subscriber do
    sequence(:stripe_customer_id) { |n| "cus_test#{n}" }
    sequence(:stripe_subscription_id) { |n| "sub_test#{n}" }
    stripe_price_id { "price_monthly" }
    stripe_status { "active" }
    stripe_current_period_start { 1.month.ago }
    stripe_current_period_end { 1.month.from_now }

    trait :active do
      stripe_status { "active" }
    end

    trait :past_due do
      stripe_status { "past_due" }
    end

    trait :canceled do
      stripe_status { "canceled" }
      stripe_canceled_at { Time.zone.now }
    end

    trait :unpaid do
      stripe_status { "unpaid" }
    end
  end
end
