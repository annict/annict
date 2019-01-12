# frozen_string_literal: true

FactoryBot.define do
  factory :status_tip, class: Tip do
    target { 0 }
    slug { "status" }
    title { "作品のステータスを変更しよう" }
    icon_name { 0 }
  end

  factory :channel_tip, class: Tip do
    target { 0 }
    slug { "channel" }
    title { "チャンネルを設定しよう" }
    icon_name { 0 }
  end

  factory :record_tip, class: Tip do
    target { 0 }
    slug { "record" }
    title { "エピソードを記録しよう" }
    icon_name { 0 }
  end
end
