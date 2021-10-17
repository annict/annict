# frozen_string_literal: true

FactoryBot.define do
  factory :gumroad_subscriber do
    gumroad_created_at { Time.zone.now }
    gumroad_product_name { "product-name" }
    gumroad_purchase_ids { [] }
    gumroad_id { "gumroad-id" }
    gumroad_product_id { "gumroad-product-id" }
  end
end
