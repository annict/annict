FactoryGirl.define do
  factory :edit_request do
    sequence(:title) { |n| "#{n}回目の編集リクエスト" }
  end
end
