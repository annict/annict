# frozen_string_literal: true

FactoryBot.define do
  factory :profile do
    sequence(:name) { |n| "人造人間#{n}号" }
    description { "悟空を倒すために生まれました。よろしくお願いします。" }
    url { "http://example.com" }
    image { File.open("#{Rails.public_path.join("images/no-image.jpg")}") }
    background_image { File.open("#{Rails.public_path.join("images/no-image.jpg")}") }
  end
end
