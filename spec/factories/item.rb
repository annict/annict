# frozen_string_literal: true

FactoryBot.define do
  factory :item do
    name { "プリキュアのDVD" }
    url { "http://amazon.co.jp" }
    tombo_image { File.open("#{Rails.root}/public/images/no_image.png") }
  end
end
