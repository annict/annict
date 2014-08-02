FactoryGirl.define do
  factory :profile do
    sequence(:name) { |n| "人造人間#{n}号" }
    description     '悟空を倒すために生まれました。よろしくお願いします。'
    avatar           File.open("#{Rails.root}/db/data/image/user_1.png")
    background_image File.open("#{Rails.root}/db/data/image/user_1.png")
  end
end
